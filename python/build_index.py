"""
build_index.py
扫描 products/ 目录，用 CLIP 提取每张图的特征向量，建立 FAISS 索引。
运行一次即可，生成 products.index 和 products.json。
"""
import os
import json
import numpy as np
import faiss
from PIL import Image
from transformers import CLIPProcessor, CLIPVisionModelWithProjection
import torch

PRODUCTS_DIR = "products"
INDEX_FILE   = "products.index"
NAMES_FILE   = "products.json"

SUPPORTED_EXT = {".jpg", ".jpeg", ".png", ".webp"}

def load_clip():
    print("加载 CLIP 模型...")
    model = CLIPVisionModelWithProjection.from_pretrained("openai/clip-vit-base-patch32")
    processor = CLIPProcessor.from_pretrained("openai/clip-vit-base-patch32")
    model.eval()
    return model, processor

def embed_image(model, processor, path):
    img = Image.open(path).convert("RGB")
    inputs = processor(images=img, return_tensors="pt")
    with torch.no_grad():
        output = model(pixel_values=inputs["pixel_values"])
    # image_embeds: (1, 512) → (512,)
    feat = output.image_embeds[0].cpu().numpy().astype("float32")
    feat /= (np.linalg.norm(feat) + 1e-8)
    return feat

def build():
    model, processor = load_clip()

    files = [
        f for f in os.listdir(PRODUCTS_DIR)
        if os.path.splitext(f)[1].lower() in SUPPORTED_EXT
    ]
    files.sort()
    print(f"共找到 {len(files)} 张商品图片")

    dim = 512  # clip-vit-base-patch32 输出维度
    index = faiss.IndexFlatIP(dim)  # 内积 = 余弦相似度（向量已归一化）

    names   = []
    vectors = []
    failed  = []

    for i, fname in enumerate(files):
        path = os.path.join(PRODUCTS_DIR, fname)
        name = os.path.splitext(fname)[0]  # 文件名去掉扩展名 = 商品名
        try:
            vec = embed_image(model, processor, path)
            vectors.append(vec)
            names.append(name)
            if (i + 1) % 50 == 0:
                print(f"  已处理 {i+1}/{len(files)}...")
        except Exception as e:
            failed.append(fname)
            print(f"  跳过 {fname}: {e}")

    index.add(np.stack(vectors).astype("float32"))
    faiss.write_index(index, INDEX_FILE)
    json.dump(names, open(NAMES_FILE, "w", encoding="utf-8"), ensure_ascii=False)

    print(f"\n完成！索引 {len(names)} 个商品 → {INDEX_FILE}")
    if failed:
        print(f"失败 {len(failed)} 个: {failed}")

if __name__ == "__main__":
    build()
