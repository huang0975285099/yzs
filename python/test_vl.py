"""
诊断 Qwen-VL API。运行: python test_vl.py
"""
import requests, base64
from PIL import Image
import io

MODEL = "qwen3-vl-8b-q4_k_m.gguf"
URL   = "http://10.0.6.226:11435/v1/chat/completions"

# 完全模拟浏览器请求头
HEADERS = {
    "Content-Type":   "application/json",
    "Accept":         "*/*",
    "Accept-Language":"en-US,en;q=0.9",
    "Origin":         "http://10.0.6.226:11435",
    "Referer":        "http://10.0.6.226:11435/",
    "User-Agent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36",
}

def test(label, payload):
    print(f"\n=== {label} ===")
    try:
        r = requests.post(URL, json=payload, headers=HEADERS, timeout=30)
        print(f"HTTP {r.status_code}")
        # stream=true 时按行读取
        if r.headers.get("content-type","").startswith("text/event-stream"):
            for line in r.iter_lines():
                if line:
                    print(line[:200])
                    break  # 只看第一行
        else:
            print(r.text[:400])
    except Exception as e:
        print(f"异常: {e}")

# 1. 纯文字（完全按浏览器格式）
test("纯文字 stream=true", {
    "model": MODEL,
    "messages": [{"role": "user", "content": "你好，只回答 OK"}],
    "stream": True,
    "return_progress": True,
    "reasoning_format": "auto",
    "backend_sampling": False,
    "timings_per_token": True,
})

# 2. 纯文字 stream=false
test("纯文字 stream=false", {
    "model": MODEL,
    "messages": [{"role": "user", "content": "你好，只回答 OK"}],
    "stream": False,
})

# 64x64 纯色测试图
img = Image.new("RGB", (64, 64), color=(255, 200, 100))
buf = io.BytesIO()
img.save(buf, format="JPEG", quality=85)
b64 = base64.b64encode(buf.getvalue()).decode()

# 3. 带图 stream=true
test("带图 stream=true", {
    "model": MODEL,
    "messages": [{
        "role": "user",
        "content": [
            {"type": "image_url", "image_url": {"url": f"data:image/jpeg;base64,{b64}"}},
            {"type": "text", "text": "图片什么颜色？只说颜色。"},
        ]
    }],
    "stream": True,
    "return_progress": True,
    "reasoning_format": "auto",
    "backend_sampling": False,
    "timings_per_token": True,
})
