# 商品识别服务

## 首次安装

```bash
cd D:\project\go-yzs\python

# 安装依赖（CPU 版 torch，约 800MB）
pip install -r requirements.txt

# 建立商品向量索引（只需运行一次，约 5-10 分钟）
python build_index.py
```

生成文件：
- `products.index`  — FAISS 向量索引
- `products.json`   — 商品名称列表

## 启动服务

```bash
python service.py
# 监听 0.0.0.0:5001
```

## 接口说明

### 健康检查
```
GET http://localhost:5001/health
```

### 上传图片搜索
```bash
curl -X POST http://localhost:5001/search/image \
  -F "image=@/path/to/test.jpg"
```

### 视频 URL 截帧搜索
```bash
curl -X POST http://localhost:5001/search/video \
  -H "Content-Type: application/json" \
  -d '{"url": "https://oss-lotus.uboxol.com/xxx.mp4?..."}'
```

返回示例：
```json
{
  "frames_used": 4,
  "results": [
    {"name": "OYee椰子水350ml", "score": 0.9213, "pct": "92.1%"},
    {"name": "PET佳果源100%NFC椰子水350ml", "score": 0.8876, "pct": "88.8%"},
    {"name": "矿泉水500ml", "score": 0.6123, "pct": "61.2%"}
  ]
}
```
