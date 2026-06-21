"""
service.py  —  Flask API
Pipeline（上传图片）：
  YOLOv8 自动裁剪商品区域 → CLIP 视觉召回 Top-15 → Qwen-VL 精确识别
"""
import os
import io
import re
import json
import base64
import numpy as np
import faiss
import requests
import cv2
from PIL import Image
from flask import Flask, request, jsonify
from transformers import CLIPProcessor, CLIPVisionModelWithProjection
from ultralytics import YOLO
import torch

INDEX_FILE = "products.index"
NAMES_FILE = "products.json"
TOP_K      = 8
CLIP_K     = 15   # 送给 Qwen-VL 的候选数量

PRODUCTS_DIR  = "products"
SUPPORTED_EXT = {".jpg", ".jpeg", ".png", ".webp"}

# Qwen-VL（llama-server OpenAI 兼容接口）
QWEN_VL_URL   = "http://10.0.6.226:11435/v1/chat/completions"
QWEN_VL_MODEL = "qwen3-vl-8b-q4_k_m.gguf"

app = Flask(__name__, static_folder="ui", static_url_path="/ui")

# ── 设备 ─────────────────────────────────────────────────────
DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
print(f"使用设备: {DEVICE}")

# ── CLIP（GPU）────────────────────────────────────────────────
print("加载 CLIP 模型...")
os.environ.setdefault("TRANSFORMERS_OFFLINE", "1")   # 优先使用本地缓存，不访问 HuggingFace
_CLIP_ID = "openai/clip-vit-base-patch32"
_clip_model     = CLIPVisionModelWithProjection.from_pretrained(
    _CLIP_ID, local_files_only=True).to(DEVICE)
_clip_processor = CLIPProcessor.from_pretrained(_CLIP_ID, local_files_only=True)
_clip_model.eval()

# ── YOLOv8（GPU，首次运行自动下载 yolov8n.pt）───────────────
print("加载 YOLOv8 模型...")
_yolo = YOLO("yolov8n.pt")
_yolo.to(DEVICE)
print("YOLOv8 就绪")

# ── FAISS + 商品名 ────────────────────────────────────────────
print("加载 FAISS 索引...")
index = faiss.read_index(INDEX_FILE)
names = json.load(open(NAMES_FILE, encoding="utf-8"))

name_to_file = {}
for f in os.listdir(PRODUCTS_DIR):
    name, ext = os.path.splitext(f)
    if ext.lower() in SUPPORTED_EXT:
        name_to_file[name] = f

print(f"就绪，共 {len(names)} 个商品")

# ── 置信度阈值（纯 CLIP 模式用） ─────────────────────────────
CONF_LOW    = 0.72
CONF_HIGH   = 0.80
CONF_SPREAD = 0.05


# ── 核心函数 ─────────────────────────────────────────────────
def embed_pil(pil_img):
    pil_img = pil_img.convert("RGB")
    inputs  = _clip_processor(images=pil_img, return_tensors="pt")
    inputs  = {k: v.to(DEVICE) for k, v in inputs.items()}
    with torch.no_grad():
        output = _clip_model(pixel_values=inputs["pixel_values"])
    feat = output.image_embeds[0].cpu().numpy().astype("float32")
    feat /= (np.linalg.norm(feat) + 1e-8)
    return feat


def pil_to_base64(pil_img, quality=90):
    buf = io.BytesIO()
    pil_img.convert("RGB").save(buf, format="JPEG", quality=quality)
    return base64.b64encode(buf.getvalue()).decode()


# COCO 中与商品相关的类别（bottle=39, cup=41, wine glass=40, bowl=45 等）
_PRODUCT_CLASSES = {39, 40, 41, 45, 46, 47}
# 永远不用作商品框的类别（人、各种车辆、动物等）
_EXCLUDE_CLASSES = {0, 1, 2, 3, 4, 5, 6, 7, 8, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23}


def _trim_dark_borders(pil_img, threshold=20):
    """裁掉图片四周的纯黑/深色边框。"""
    arr = np.array(pil_img.convert("RGB"))
    gray = arr.mean(axis=2)  # H×W
    col_bright = gray.mean(axis=0)   # 每列亮度均值
    row_bright = gray.mean(axis=1)   # 每行亮度均值
    cols = np.where(col_bright > threshold)[0]
    rows = np.where(row_bright > threshold)[0]
    if len(cols) == 0 or len(rows) == 0:
        return pil_img
    x1, x2 = int(cols[0]), int(cols[-1]) + 1
    y1, y2 = int(rows[0]), int(rows[-1]) + 1
    if (x2 - x1) < pil_img.width * 0.5 or (y2 - y1) < pil_img.height * 0.5:
        return pil_img  # 裁掉太多，放弃
    trimmed = pil_img.crop((x1, y1, x2, y2))
    print(f"[TRIM] 去黑边: ({x1},{y1})-({x2},{y2})")
    return trimmed


def _yolo_best_box(pil_img, scale=1.0):
    """在给定图上跑 YOLO，返回最佳商品框坐标（原图空间）或 None。"""
    W, H = pil_img.width, pil_img.height
    detect_img = pil_img
    s = scale
    if min(W, H) < 640:
        s = scale * 640 / min(W, H)
        detect_img = pil_img.resize((int(W * s / scale), int(H * s / scale)), Image.LANCZOS)

    arr = np.array(detect_img.convert("RGB"))
    results = _yolo(arr, verbose=False, conf=0.25, device=DEVICE)
    if not results or len(results[0].boxes) == 0:
        return None

    dW, dH = detect_img.width, detect_img.height
    best_box, best_score = None, -1.0
    for box in results[0].boxes:
        cls  = int(box.cls[0])
        if cls in _EXCLUDE_CLASSES:
            continue
        conf = float(box.conf[0])
        bx1 = int(box.xyxy[0][0]); by1 = int(box.xyxy[0][1])
        bx2 = int(box.xyxy[0][2]); by2 = int(box.xyxy[0][3])
        area_ratio = (bx2 - bx1) * (by2 - by1) / (dW * dH)
        min_area = 0.005 if cls in _PRODUCT_CLASSES else 0.05   # 商品类 0.5%
        if area_ratio < min_area or area_ratio > 0.92:
            continue
        score = conf * (3.0 if cls in _PRODUCT_CLASSES else 1.0)
        if score > best_score:
            best_score = score
            # 映射回原图坐标
            rx1 = int(bx1 * W / dW); ry1 = int(by1 * H / dH)
            rx2 = int(bx2 * W / dW); ry2 = int(by2 * H / dH)
            best_box = (rx1, ry1, rx2, ry2)
    return best_box


def _crop_with_pad(pil_img, box, pad=0.08):
    """按 box 裁剪并加 pad 边距。"""
    W, H = pil_img.width, pil_img.height
    x1, y1, x2, y2 = box
    px = int((x2 - x1) * pad)
    py = int((y2 - y1) * pad)
    x1 = max(0, x1 - px);  y1 = max(0, y1 - py)
    x2 = min(W, x2 + px);  y2 = min(H, y2 + py)
    return pil_img.crop((x1, y1, x2, y2)), (x1, y1, x2, y2)


def detect_and_crop(pil_img):
    """
    商品定位裁剪，三级策略：
      1. Qwen-VL grounding —— 语义定位，袋装/盒装/瓶装通吃（首选）
      2. YOLOv8 —— COCO 类别（bottle/cup 等）兜底
      3. 去黑边 + 中心裁剪 —— 最后兜底
    """
    W, H = pil_img.width, pil_img.height

    # 策略 1：VL 语义定位
    vl_box = qwen_vl_locate(pil_img)
    if vl_box:
        cropped, box = _crop_with_pad(pil_img, vl_box, pad=0.06)
        print(f"[CROP] VL定位裁剪: {box}")
        return cropped, box

    # 策略 2：YOLO（宽图先检测中间列）
    print("[CROP] VL未定位，尝试 YOLO")
    search_img = pil_img
    offset_x = 0
    if W / H > 2.5:
        x_start = W // 3
        x_end   = W * 2 // 3
        search_img = pil_img.crop((x_start, 0, x_end, H))
        offset_x = x_start
        print(f"[YOLO] 宽图，先检测中间列 ({x_start},{x_end})")

    best_box = _yolo_best_box(search_img)
    if best_box is None and offset_x > 0:
        print("[YOLO] 中间列未检到，全图再试")
        best_box = _yolo_best_box(pil_img)
        offset_x = 0

    if best_box:
        x1, y1, x2, y2 = best_box
        cropped, box = _crop_with_pad(pil_img,
            (x1 + offset_x, y1, x2 + offset_x, y2), pad=0.08)
        print(f"[YOLO] 裁剪: {box}")
        return cropped, box

    # 策略 3：兜底
    print("[CROP] 全部失败，走去黑边+中心裁剪")
    return _fallback_crop(pil_img), None


def _fallback_crop(pil_img):
    """YOLO 无结果时的降级处理：去黑边 → 中心 80% 裁剪。"""
    img = _trim_dark_borders(pil_img)
    W, H = img.size
    # 保留中心 80%，去掉四周杂乱背景
    margin_x = int(W * 0.10)
    margin_y = int(H * 0.10)
    cropped = img.crop((margin_x, margin_y, W - margin_x, H - margin_y))
    print(f"[FALLBACK] 中心裁剪后尺寸: {cropped.size}")
    return cropped


def _fuzzy_match_name(answer, name_list, threshold=0.5):
    """双向 F1 汉字匹配：兼顾精确率（answer 的字在商品名里）和召回率（商品名的字在 answer 里）。"""
    answer_chars = set(ch for ch in answer if '一' <= ch <= '鿿')
    if not answer_chars:
        return (None, 0.0)
    best_name, best_score = None, 0.0
    for name in name_list:
        name_chars = set(ch for ch in name if '一' <= ch <= '鿿')
        if not name_chars:
            continue
        inter = len(answer_chars & name_chars)
        if inter == 0:
            continue
        precision = inter / len(answer_chars)   # answer 里的字有多少在商品名里
        recall    = inter / len(name_chars)      # 商品名里的字有多少在 answer 里
        f1 = 2 * precision * recall / (precision + recall)
        if f1 > best_score:
            best_score, best_name = f1, name
    return (best_name, best_score) if best_score >= threshold else (None, 0.0)


def _vl_call(pil_img, prompt, max_tokens=80):
    """发送 VL 请求，返回回答字符串；失败返回 None。"""
    img = pil_img.convert("RGB")
    w, h = img.size
    if max(w, h) > 640:
        scale = 640 / max(w, h)
        img = img.resize((int(w * scale), int(h * scale)), Image.LANCZOS)
    b64 = pil_to_base64(img, quality=85)
    print(f"[VL] 图片: {img.size}, base64: {len(b64)} bytes")
    payload = {
        "model": "local",
        "messages": [{"role": "user", "content": [
            {"type": "text", "text": prompt},
            {"type": "image_url", "image_url": {"url": f"data:image/jpeg;base64,{b64}"}},
        ]}],
        "temperature": 0.1,
        "max_tokens": max_tokens,
        "stream": False,
    }
    resp = requests.post(QWEN_VL_URL, json=payload, timeout=60)
    print(f"[VL] HTTP {resp.status_code}")
    if resp.status_code != 200:
        print(f"[VL] 错误: {resp.text[:300]}")
        return None
    return resp.json()["choices"][0]["message"]["content"].strip()


def qwen_vl_locate(pil_img):
    """
    用 Qwen-VL 做目标定位（grounding），框出画面中的商品。
    COCO YOLO 检测不到袋装/盒装零食，VL 靠语义理解能定位任意商品。
    返回原图坐标 (x1, y1, x2, y2) 或 None。
    """
    try:
        W, H = pil_img.size
        # 长边缩到 1000，保证定位分辨率，同时记录发送尺寸用于坐标映射
        long = max(W, H)
        s = 1000 / long if long > 1000 else 1.0
        sent = pil_img.convert("RGB")
        if s != 1.0:
            sent = sent.resize((int(W * s), int(H * s)), Image.LANCZOS)
        sw, sh = sent.size
        b64 = pil_to_base64(sent, quality=85)

        prompt = (
            "图片中有人手持一件商品（饮料瓶、零食袋、盒装食品等）。"
            "请定位这件商品包装的位置，输出它的边界框像素坐标。"
            f"图片宽{sw}像素、高{sh}像素。"
            "只输出一个坐标数组，格式：[左上x,左上y,右下x,右下y]，不要任何其他文字。"
        )
        payload = {
            "model": "local",
            "messages": [{"role": "user", "content": [
                {"type": "text", "text": prompt},
                {"type": "image_url", "image_url": {"url": f"data:image/jpeg;base64,{b64}"}},
            ]}],
            "temperature": 0.0,
            "max_tokens": 40,
            "stream": False,
        }
        resp = requests.post(QWEN_VL_URL, json=payload, timeout=60)
        if resp.status_code != 200:
            print(f"[VL-LOC] HTTP {resp.status_code}: {resp.text[:150]}")
            return None
        answer = resp.json()["choices"][0]["message"]["content"].strip()
        print(f"[VL-LOC] 原始坐标: {answer!r} (发送尺寸 {sw}x{sh})")

        nums = re.findall(r"-?\d+\.?\d*", answer)
        if len(nums) < 4:
            print("[VL-LOC] 坐标数量不足")
            return None
        x1, y1, x2, y2 = [float(n) for n in nums[:4]]

        # 判断坐标系：像素 / 0-1000 归一化 / 0-1 归一化
        mx = max(x1, y1, x2, y2)
        if mx <= 1.5:                              # 0-1 归一化
            fx1, fy1, fx2, fy2 = x1, y1, x2, y2
        elif x1 <= sw and x2 <= sw and y1 <= sh and y2 <= sh:  # 像素坐标
            fx1, fy1, fx2, fy2 = x1/sw, y1/sh, x2/sw, y2/sh
        else:                                       # 0-1000 归一化
            fx1, fy1, fx2, fy2 = x1/1000, y1/1000, x2/1000, y2/1000

        # 映射回原图
        rx1, ry1 = int(fx1 * W), int(fy1 * H)
        rx2, ry2 = int(fx2 * W), int(fy2 * H)
        rx1, rx2 = sorted((max(0, rx1), min(W, rx2)))
        ry1, ry2 = sorted((max(0, ry1), min(H, ry2)))

        bw, bh = rx2 - rx1, ry2 - ry1
        area_ratio = (bw * bh) / (W * H)
        if bw < 10 or bh < 10 or area_ratio < 0.005 or area_ratio > 0.95:
            print(f"[VL-LOC] 框无效 占比{area_ratio:.1%}")
            return None
        print(f"[VL-LOC] 定位: ({rx1},{ry1},{rx2},{ry2}) 占比{area_ratio:.1%}")
        return (rx1, ry1, rx2, ry2)

    except Exception as e:
        print(f"[VL-LOC] 失败: {e}")
        return None


def qwen_vl_identify(pil_img):
    """VL 自由识别：直接读标签，返回商品名称（不限候选）。"""
    prompt = (
        "请识别图片中的商品，按格式回答：\n"
        "商品名称（品牌+产品名+口味+规格）|包装类型|主色调\n"
        "包装类型从以下选择：罐装、瓶装、袋装、盒装、桶装、其他\n"
        "主色调用1-2个颜色词描述。\n"
        "例如：'康师傅冰红茶蜜桃味500ml|瓶装|粉红色'\n"
        "如果信息不全，只写能看到的部分，不要猜测，不要解释。"
    )
    answer = _vl_call(pil_img, prompt, max_tokens=60)
    if answer:
        print(f"[VL-FREE] {answer!r}")
    return answer


def _clip_search(pil_img):
    """CLIP 4 角度搜索，返回按分数排序的候选列表。"""
    clip_best = {}
    for angle in [0, 90, 180, 270]:
        img  = pil_img.rotate(angle, expand=True) if angle else pil_img
        feat = embed_pil(img)
        scores, indices = index.search(feat.reshape(1, -1), CLIP_K)
        for s, idx in zip(scores[0], indices[0]):
            if idx >= 0:
                n = names[idx]
                if n not in clip_best or float(s) > clip_best[n]:
                    clip_best[n] = float(s)
    return sorted(
        [{"name": n, "score": round(s, 4), "clip_score": round(s, 4),
          "pct": f"{s*100:.1f}%",
          "img": f"/products/{name_to_file[n]}" if n in name_to_file else ""}
         for n, s in clip_best.items()],
        key=lambda x: -x["score"]
    )


def qwen_vl_visual_select(pil_query, candidates, top_n=5):
    """
    让 VL 同时看查询图 + 候选商品图，视觉比对选出最匹配的。
    比纯文字候选列表更准确。
    """
    try:
        # 查询图（缩到短边不超过 400）
        q = pil_query.convert("RGB")
        if max(q.size) > 400:
            s = 400 / max(q.size)
            q = q.resize((int(q.width * s), int(q.height * s)), Image.LANCZOS)
        q_b64 = pil_to_base64(q, quality=80)

        # 加载候选商品图
        valid = []
        for c in candidates[:top_n]:
            fname = name_to_file.get(c["name"], "")
            if not fname:
                continue
            path = os.path.join(PRODUCTS_DIR, fname)
            if not os.path.exists(path):
                continue
            try:
                ci = Image.open(path).convert("RGB")
                if max(ci.size) > 256:
                    s = 256 / max(ci.size)
                    ci = ci.resize((int(ci.width * s), int(ci.height * s)), Image.LANCZOS)
                valid.append({"name": c["name"], "b64": pil_to_base64(ci, quality=75)})
            except Exception:
                continue

        if not valid:
            return None

        # 构建多图消息：查询图 + 候选图并排
        name_list = "\n".join(f"{i+1}. {c['name']}" for i, c in enumerate(valid))
        content = [
            {"type": "text",      "text": "这是需要识别的商品图（查询图）："},
            {"type": "image_url", "image_url": {"url": f"data:image/jpeg;base64,{q_b64}"}},
            {"type": "text",      "text": f"下面是 {len(valid)} 张候选商品图，顺序对应以下名称：\n{name_list}\n\n请对比查询图，只回答最匹配的商品名称。"},
        ]
        for c in valid:
            content.append({"type": "image_url",
                             "image_url": {"url": f"data:image/jpeg;base64,{c['b64']}"}})

        payload = {
            "model": "local",
            "messages": [{"role": "user", "content": content}],
            "temperature": 0.1,
            "max_tokens": 60,
            "stream": False,
        }
        resp = requests.post(QWEN_VL_URL, json=payload, timeout=120)
        print(f"[VL-VIS] HTTP {resp.status_code}")
        if resp.status_code != 200:
            print(f"[VL-VIS] 错误: {resp.text[:200]}")
            return None
        answer = resp.json()["choices"][0]["message"]["content"].strip()
        print(f"[VL-VIS] 回答: {answer!r}")

        # 精确匹配
        vnames = [c["name"] for c in valid]
        for name in vnames:
            if name in answer or answer in name:
                return name
        # F1 模糊匹配
        best, _ = _fuzzy_match_name(answer, vnames, threshold=0.4)
        return best

    except Exception as e:
        print(f"[VL-VIS] 失败: {e}")
        return None


def qwen_vl_select(pil_img, candidates):
    """
    用 Qwen-VL 识别商品：
    1. 先让 VL 从 CLIP 候选列表中选；
    2. 候选里找不到时，让 VL 自由回答商品名，再全库模糊匹配。
    返回商品名（str）或 None。
    """
    try:
        img = pil_img.convert("RGB")
        w, h = img.size
        if max(w, h) > 640:
            scale = 640 / max(w, h)
            img = img.resize((int(w * scale), int(h * scale)), Image.LANCZOS)

        b64 = pil_to_base64(img, quality=85)
        print(f"[VL] 图片: {img.size}, base64: {len(b64)} bytes")

        cands_text = "\n".join(f"{i+1}. {c['name']}" for i, c in enumerate(candidates))
        prompt = (
            "请看这张图片中的商品。\n"
            "从以下列表中选出最匹配的商品名称，只回答名称本身，不加序号；\n"
            "如果列表中没有匹配项，请直接回答你识别到的商品名称（品牌+产品名）：\n"
            + cands_text
        )
        payload = {
            "model":       "local",
            "messages":    [{
                "role": "user",
                "content": [
                    {"type": "text", "text": prompt},
                    {"type": "image_url",
                     "image_url": {"url": f"data:image/jpeg;base64,{b64}"}},
                ]
            }],
            "temperature": 0.1,
            "max_tokens":  80,
            "stream":      False,
        }
        resp = requests.post(QWEN_VL_URL, json=payload, timeout=60)
        print(f"[VL] HTTP {resp.status_code}")
        if resp.status_code != 200:
            print(f"[VL] 错误: {resp.text[:300]}")
            resp.raise_for_status()
        answer = resp.json()["choices"][0]["message"]["content"].strip()
        print(f"[VL] 回答: {answer!r}")

        # Step 1: 精确匹配候选
        for c in candidates:
            if c["name"] in answer or answer in c["name"]:
                return c["name"]

        # Step 2: 候选列表里模糊匹配
        cand_names = [c["name"] for c in candidates]
        best_name, best_score = _fuzzy_match_name(answer, cand_names, threshold=0.5)
        if best_name:
            print(f"[VL] 候选模糊匹配: {best_name} ({best_score:.2f})")
            return best_name

        # Step 3: 全库模糊匹配（VL 自由回答了真实商品名）
        print(f"[VL] 候选未匹配，全库搜索: {answer!r}")
        best_name, best_score = _fuzzy_match_name(answer, names, threshold=0.4)
        if best_name:
            print(f"[VL] 全库模糊匹配: {best_name} ({best_score:.2f})")
            return best_name

        print(f"[VL] 全库也未匹配: {answer!r}")
        return None

    except Exception as e:
        print(f"[VL] 调用失败: {e}")
        return None


def search(feat, top_k=TOP_K):
    """纯 CLIP 搜索（视频/截帧接口用），含置信度警告。"""
    scores, indices = index.search(feat.reshape(1, -1), top_k)
    results = []
    for score, idx in zip(scores[0], indices[0]):
        if idx >= 0:
            name     = names[idx]
            img_file = name_to_file.get(name, "")
            results.append({
                "name":  name,
                "score": round(float(score), 4),
                "pct":   f"{float(score)*100:.1f}%",
                "img":   f"/products/{img_file}" if img_file else "",
            })
    warn = ""
    if results:
        top_score = results[0]["score"]
        if top_score < CONF_LOW:
            warn = f"最高分 {top_score*100:.1f}% 过低，图像不清晰或商品不在库中"
        elif top_score < CONF_HIGH and len(results) > 1 \
                and (top_score - results[1]["score"]) < CONF_SPREAD:
            warn = f"Top1({top_score*100:.1f}%) 与 Top2({results[1]['score']*100:.1f}%) 接近，请仔细确认"
    if warn:
        results.insert(0, {"__warn__": warn})
    return results


def merge_results(all_results):
    """多帧结果合并：同一商品取最高分，跳过警告项。"""
    best = {}
    for r in all_results:
        if "name" not in r:
            continue
        name = r["name"]
        if name not in best or r["score"] > best[name]["score"]:
            best[name] = r
    return sorted(best.values(), key=lambda x: x["score"], reverse=True)[:TOP_K]


def _build_result_item(name, score=1.0, pct="VL识别", vl_match=False):
    img_file = name_to_file.get(name, "")
    return {
        "name": name, "score": score, "clip_score": 0.0,
        "pct": pct, "vl_match": vl_match,
        "img": f"/products/{img_file}" if img_file else "",
    }


def _promote_vl(results, vl_choice):
    """把 vl_choice 提到结果第一位，如不在列表里则插入。"""
    vl_item = next((r for r in results if r["name"] == vl_choice), None)
    if vl_item:
        vl_item["vl_match"] = True
        return [vl_item] + [r for r in results if r["name"] != vl_choice]
    results.insert(0, _build_result_item(vl_choice, vl_match=True))
    return results[:TOP_K]


def search_hybrid(pil_img):
    """
    VL 优先架构（不裁剪，VL 直接看全图）：
    Step 1: 去黑边 → VL 读文字/品牌 → 全库匹配，置信则直接返回
    Step 2: 读不出/匹配不上 → CLIP 召回 → VL 看候选图挑最像的一张
    返回 (results, vl_answer, preview_b64)。
    """
    # 预处理：只去黑边，不做破坏性裁剪
    pil_clean = _trim_dark_borders(pil_img)
    preview_b64 = pil_to_base64(pil_clean.resize(
        (min(pil_clean.width, 400),
         int(pil_clean.height * min(400 / pil_clean.width, 1))),
        Image.LANCZOS), quality=80)

    # Step 1: VL 读文字（回答格式：名称|包装类型|颜色）
    vl_free = qwen_vl_identify(pil_clean)
    vl_name_part = vl_free.split("|")[0].strip() if vl_free else ""
    if vl_name_part:
        # 阈值 0.6，避免「牛肉干→牛肉面」这类近似误匹配；不确定的交给视觉比对
        best_name, best_score = _fuzzy_match_name(vl_name_part, names, threshold=0.6)
        if best_name:
            print(f"[HYBRID] VL文字命中: {best_name} ({best_score:.2f})")
            clip_results = _clip_search(pil_clean)
            results = _promote_vl(clip_results[:TOP_K], best_name)
            return results, best_name, preview_b64

    # Step 2: CLIP 召回 → VL 看候选图挑最像的
    print(f"[HYBRID] 文字未确信匹配（VL读到:{vl_name_part!r}），CLIP召回+VL视觉比对")
    candidates = _clip_search(pil_clean)
    vl_choice  = qwen_vl_visual_select(pil_clean, candidates)        # 发真实商品图给 VL 比对
    if not vl_choice:
        vl_choice = qwen_vl_select(pil_clean, candidates[:CLIP_K])  # 降级：纯文字候选列表
    vl_answer  = vl_choice or (vl_name_part or "")

    results = candidates[:TOP_K]
    if vl_choice:
        results = _promote_vl(results, vl_choice)

    return results, vl_answer, preview_b64


# ── 接口一：上传图片（YOLOv8 + CLIP + Qwen-VL）──────────────
@app.route("/search/image", methods=["POST"])
def search_by_image():
    try:
        if request.content_type and "multipart" in request.content_type:
            f = request.files.get("image")
            if not f:
                return jsonify({"error": "缺少 image 字段"}), 400
            pil_img = Image.open(f.stream)
        else:
            data      = request.get_json(force=True)
            img_bytes = base64.b64decode(data["image"])
            pil_img   = Image.open(io.BytesIO(img_bytes))

        results, vl_answer, cropped_b64 = search_hybrid(pil_img)
        return jsonify({"results": results, "vl_answer": vl_answer,
                        "cropped_img": "data:image/jpeg;base64," + cropped_b64})

    except Exception as e:
        return jsonify({"error": str(e)}), 500


# ── 接口二：视频 URL 截帧搜索（纯 CLIP）─────────────────────────
@app.route("/search/video", methods=["POST"])
def search_by_video():
    try:
        data      = request.get_json(force=True)
        video_url = data.get("url", "").strip()
        n_frames  = int(data.get("n_frames", 4))
        if not video_url:
            return jsonify({"error": "缺少 url 字段"}), 400

        frames = extract_frames_from_url(video_url, n_frames=n_frames)
        if not frames:
            return jsonify({"error": "截帧失败，检查视频 URL 是否有效"}), 400

        all_results = []
        for img in frames:
            feat = embed_pil(img)
            all_results.extend(search(feat, top_k=TOP_K))

        merged = merge_results(all_results)
        return jsonify({"frames_used": len(frames), "results": merged})

    except Exception as e:
        return jsonify({"error": str(e)}), 500


# ── 接口三：截帧返回 base64 ──────────────────────────────────
@app.route("/capture/frame", methods=["POST"])
def capture_frame():
    try:
        data      = request.get_json(force=True)
        video_url = data.get("url", "").strip()
        timestamp = float(data.get("timestamp", 0))
        if not video_url:
            return jsonify({"error": "缺少 url 字段"}), 400

        pil_img = capture_frame_pil(video_url, timestamp)
        buf = io.BytesIO()
        pil_img.save(buf, format="JPEG", quality=85)
        b64 = base64.b64encode(buf.getvalue()).decode()
        return jsonify({
            "image": b64, "width": pil_img.width,
            "height": pil_img.height, "timestamp": timestamp,
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500


# ── 接口四：裁剪区域识别 ─────────────────────────────────────
@app.route("/search/crop", methods=["POST"])
def search_by_crop():
    try:
        data      = request.get_json(force=True)
        video_url = data.get("url", "").strip()
        timestamp = float(data.get("timestamp", 0))
        crop      = data.get("crop")
        if not video_url:
            return jsonify({"error": "缺少 url 字段"}), 400

        pil_img = capture_frame_pil(video_url, timestamp)
        W, H    = pil_img.width, pil_img.height
        if crop:
            x  = int(crop["x"] * W); y = int(crop["y"] * H)
            w  = int(crop["w"] * W); h = int(crop["h"] * H)
            x2 = min(x + w, W);     y2 = min(y + h, H)
            if x2 > x and y2 > y:
                pil_img = pil_img.crop((x, y, x2, y2))

        feat    = embed_pil(pil_img)
        results = search(feat)
        return jsonify({"timestamp": timestamp, "results": results})

    except Exception as e:
        return jsonify({"error": str(e)}), 500


# ── 接口五：直接截帧识别（兼容旧版）─────────────────────────────
@app.route("/search/frame", methods=["POST"])
def search_by_frame():
    try:
        data      = request.get_json(force=True)
        video_url = data.get("url", "").strip()
        timestamp = float(data.get("timestamp", 0))
        if not video_url:
            return jsonify({"error": "缺少 url 字段"}), 400
        pil_img = capture_frame_pil(video_url, timestamp)
        feat    = embed_pil(pil_img)
        results = search(feat)
        return jsonify({"timestamp": timestamp, "results": results})
    except Exception as e:
        return jsonify({"error": str(e)}), 500


# ── 工具函数 ─────────────────────────────────────────────────
def extract_frames_from_url(video_url, n_frames=4):
    cap = cv2.VideoCapture(video_url)
    if not cap.isOpened():
        return []
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    fps          = cap.get(cv2.CAP_PROP_FPS) or 25
    if total_frames <= 0:
        ret, _ = cap.read()
        if not ret:
            cap.release()
            return []
        total_frames = int(fps * 60)
    start     = int(total_frames * 0.2)
    end       = int(total_frames * 0.8)
    positions = [int(start + (end - start) * i / max(n_frames - 1, 1)) for i in range(n_frames)]
    frames = []
    for pos in positions:
        cap.set(cv2.CAP_PROP_POS_FRAMES, pos)
        ret, frame = cap.read()
        if ret:
            frames.append(Image.fromarray(cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)))
    cap.release()
    return frames


def capture_frame_pil(video_url, timestamp):
    cap = cv2.VideoCapture(video_url)
    if not cap.isOpened():
        raise RuntimeError("无法打开视频")
    fps = cap.get(cv2.CAP_PROP_FPS) or 25
    cap.set(cv2.CAP_PROP_POS_FRAMES, int(timestamp * fps))
    ret, frame = cap.read()
    cap.release()
    if not ret:
        raise RuntimeError(f"无法读取 {timestamp}s 处的帧")
    return Image.fromarray(cv2.cvtColor(frame, cv2.COLOR_BGR2RGB))


# ── 商品图片静态服务 ──────────────────────────────────────────
@app.route("/products/<path:filename>")
def serve_product_img(filename):
    from flask import send_from_directory
    return send_from_directory(PRODUCTS_DIR, filename)


@app.route("/")
def main_page():
    return app.send_static_file("index.html")


@app.route("/health")
def health():
    vl_ok = False
    try:
        r = requests.get(QWEN_VL_URL.replace("/v1/chat/completions", "/health"), timeout=3)
        vl_ok = r.status_code < 500
    except Exception:
        pass
    return jsonify({
        "status": "ok", "products": len(names),
        "device": DEVICE, "qwen_vl": vl_ok,
    })


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5001, debug=False)
