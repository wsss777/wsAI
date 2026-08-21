package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"wsai/backend/config"
)

type Hit struct {
	ChunkID    string
	DocumentID string
	FileName   string
	Score      float64
}

func request(ctx context.Context, method, url string, body, out interface{}) error {
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode/100 != 2 {
		x, _ := io.ReadAll(r.Body)
		return fmt.Errorf("请求 Qdrant 失败: %s", x)
	}
	return json.NewDecoder(r.Body).Decode(out)
}

func post(ctx context.Context, url string, body, out interface{}) error {
	return request(ctx, http.MethodPost, url, body, out)
}
func Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("embedding input must not be empty")
	}

	base := strings.TrimRight(os.Getenv("ZHIPU_BASE_URL"), "/")
	apiKey := os.Getenv("ZHIPU_API_KEY")
	if base == "" || apiKey == "" {
		return nil, fmt.Errorf("ZHIPU_BASE_URL and ZHIPU_API_KEY must be configured for document retrieval")
	}
	endpoint := base
	if !strings.HasSuffix(endpoint, "/embeddings") {
		endpoint += "/embeddings"
	}
	var r struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Error interface{} `json:"error"`
	}
	b := map[string]interface{}{
		"model":      EmbeddingModel(),
		"input":      texts,
		"dimensions": config.C.RAGConfig.EmbeddingSize,
	}
	raw, _ := json.Marshal(b)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if e != nil {
		return nil, e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	res, e := http.DefaultClient.Do(req)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		x, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("生成向量失败: %s", x)
	}
	if e = json.NewDecoder(res.Body).Decode(&r); e != nil {
		return nil, e
	}
	out := make([][]float64, len(r.Data))
	for i := range r.Data {
		out[i] = r.Data[i].Embedding
	}
	return out, nil
}

// EmbeddingModel 返回文档向量实际使用的模型。
// 对话与向量化有意使用独立的服务提供方和凭据。
func EmbeddingModel() string {
	if model := strings.TrimSpace(os.Getenv("ZHIPU_EMBEDDING_MODEL")); model != "" {
		return model
	}
	return config.C.RAGConfig.EmbeddingModel
}
func Ensure(ctx context.Context, size int) error {
	u := strings.TrimRight(config.C.RAGConfig.QdrantURL, "/") + "/collections/" + config.C.RAGConfig.Collection
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	r, e := http.DefaultClient.Do(req)
	if e == nil && r.StatusCode == 200 {
		r.Body.Close()
		return nil
	}
	if r != nil {
		r.Body.Close()
	}
	b := map[string]interface{}{"vectors": map[string]interface{}{"size": size, "distance": "Cosine"}}
	raw, _ := json.Marshal(b)
	req, e = http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(raw))
	if e != nil {
		return e
	}
	req.Header.Set("Content-Type", "application/json")
	r, e = http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer r.Body.Close()
	if r.StatusCode/100 != 2 && r.StatusCode != 409 {
		return fmt.Errorf("创建 Qdrant 集合失败: %s", r.Status)
	}
	return nil
}
func Upsert(ctx context.Context, points []map[string]interface{}) error {
	var out interface{}
	// Qdrant 将 POST /points 视为不同操作。写入点位必须使用 PUT，
	// 否则其解析器会期待带有 `ids` 的点位选择器。
	return request(ctx, http.MethodPut, strings.TrimRight(config.C.RAGConfig.QdrantURL, "/")+"/collections/"+config.C.RAGConfig.Collection+"/points?wait=true", map[string]interface{}{"points": points}, &out)
}
func Search(ctx context.Context, vector []float64, user string, docs []string, limit int) ([]Hit, error) {
	must := []interface{}{map[string]interface{}{"key": "user_name", "match": map[string]interface{}{"value": user}}}
	if len(docs) > 0 {
		must = append(must, map[string]interface{}{"key": "document_id", "match": map[string]interface{}{"any": docs}})
	}
	var out struct {
		Result []struct {
			Score   float64                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}
	e := post(ctx, strings.TrimRight(config.C.RAGConfig.QdrantURL, "/")+"/collections/"+config.C.RAGConfig.Collection+"/points/search", map[string]interface{}{"vector": vector, "limit": limit, "with_payload": true, "filter": map[string]interface{}{"must": must}}, &out)
	hits := make([]Hit, 0, len(out.Result))
	for _, x := range out.Result {
		hits = append(hits, Hit{ChunkID: fmt.Sprint(x.Payload["chunk_id"]), DocumentID: fmt.Sprint(x.Payload["document_id"]), FileName: fmt.Sprint(x.Payload["file_name"]), Score: x.Score})
	}
	return hits, e
}
