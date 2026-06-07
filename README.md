# mini-imageflux-go

mini-imageflux-go は、Goで実装する小規模な画像変換プロキシです。

指定されたオリジン画像を取得し、リサイズや形式変換を行ったうえで、変換済み画像をキャッシュして返します。  
画像配信基盤に必要なキャッシュ設計、HTTPキャッシュヘッダ、メトリクス計測、変換処理の負荷管理を検証するためのプロジェクトです。

## Goals

このプロジェクトでは、以下の要素を小規模に実装・検証します。

- オリジン画像を取得する
- リクエストパラメータに応じて画像を変換する
- 同一条件の変換結果をキャッシュする
- HTTPキャッシュヘッダを付与する
- レイテンシ、エラー率、キャッシュヒット率を計測する
- 高負荷時や異常入力時の課題を把握する

## Motivation

画像配信では、単に元画像をそのまま返すだけではなく、用途に応じたリサイズ、形式変換、キャッシュ制御、変換コストの管理、メトリクス計測が重要になります。

特に画像変換を伴う配信では、以下のような設計上の論点があります。

- 同じ変換処理を何度も実行しないためのキャッシュ設計
- 変換条件をどのようにキャッシュキーへ反映するか
- ブラウザやCDNに対してどのようなキャッシュヘッダを返すか
- オリジン取得、画像デコード、リサイズ、エンコードのどこがボトルネックになるか
- 巨大画像や不正なURLに対してどう制限をかけるか
- レイテンシやエラー率をどのように観測するか

このプロジェクトでは、これらの要素を段階的に実装し、画像変換プロキシに必要な基本構成を検証します。

## Features

現在の実装状況です。

- [ ] Go による HTTP サーバー
- [ ] URL 指定によるオリジン画像取得
- [ ] 画像リサイズ
- [ ] JPEG 出力
- [ ] WebP 出力
- [ ] ディスクキャッシュ
- [ ] `Cache-Control` ヘッダ
- [ ] `ETag` 対応
- [ ] `/metrics` エンドポイント
- [ ] キャッシュヒット数・ミス数の計測
- [ ] レスポンス時間の計測
- [ ] ベンチマーク結果の記録
- [ ] nginx reverse proxy 構成
- [ ] URL 署名
- [ ] 巨大画像対策

## Architecture

基本構成は以下の通りです。

```txt
Client
  |
  | HTTP Request
  v
Go Image Proxy
  |
  | 1. Generate cache key
  | 2. Check converted image cache
  | 3. Fetch origin image if cache miss
  | 4. Decode image
  | 5. Resize / convert format
  | 6. Store converted image to cache
  | 7. Return response with cache headers
  v
Disk Cache
  |
  v
Origin Image Server
```

将来的には、前段に nginx を置き、reverse proxy、proxy cache、rate limit なども検証する予定です。

```txt
Client
  |
  v
nginx
  |
  v
Go Image Proxy
  |
  v
Disk Cache / Origin
```

## Request Flow

画像変換リクエストは、以下の流れで処理します。

1. クライアントが変換条件付きでリクエストを送信する
2. リクエストパラメータを検証する
3. オリジンURL、出力サイズ、出力形式などからキャッシュキーを生成する
4. 変換済み画像がキャッシュに存在する場合は、それを返す
5. キャッシュが存在しない場合は、オリジン画像を取得する
6. 画像をデコードする
7. 指定された条件に従ってリサイズ・形式変換を行う
8. 変換後の画像をキャッシュに保存する
9. HTTPキャッシュヘッダを付与してレスポンスを返す
10. 処理時間、キャッシュヒット率、エラー数などをメトリクスに記録する

## API

### Image conversion

```txt
GET /image?url={origin_url}&w={width}&format={format}
```

### Example

```txt
GET /image?url=https://example.com/sample.jpg&w=512&format=webp
```

### Parameters

| Parameter | Description | Example |
|---|---|---|
| `url` | オリジン画像のURL | `https://example.com/sample.jpg` |
| `w` | 出力画像の横幅 | `512` |
| `format` | 出力形式 | `jpeg`, `webp` |

## Cache Strategy

変換済み画像は、同一条件のリクエストで再変換を行わないようにディスクキャッシュへ保存します。

キャッシュキーは以下の要素から生成します。

- オリジン画像URL
- 出力幅
- 出力形式
- 品質設定

例:

```txt
sha256(origin_url + width + format + quality)
```

同じオリジン画像であっても、変換条件が異なる場合は別のキャッシュとして扱います。

### Cache Hit

キャッシュヒット時は、オリジン画像の取得、画像デコード、リサイズ、エンコードを行わず、保存済みの変換結果を返します。

### Cache Miss

キャッシュミス時は、オリジン画像を取得し、指定された条件で変換を行った後、変換済み画像をキャッシュへ保存します。

## HTTP Cache Headers

変換結果には、ブラウザやCDNでのキャッシュを想定してHTTPキャッシュヘッダを付与します。

想定しているヘッダ例:

```txt
Cache-Control: public, max-age=31536000, immutable
ETag: "{hash}"
```

変換条件がURLに含まれる設計であれば、同一URLに対するレスポンスは基本的に不変として扱えます。  
そのため、長めの `max-age` や `immutable` の利用を検討します。

## Metrics

`/metrics` エンドポイントで、以下のような値を確認できるようにします。

- 総リクエスト数
- 変換成功数
- 変換失敗数
- キャッシュヒット数
- キャッシュミス数
- キャッシュヒット率
- レスポンス時間
- オリジン取得時間
- 画像変換時間

出力例:

```txt
image_proxy_requests_total 120
image_proxy_cache_hits_total 85
image_proxy_cache_misses_total 35
image_proxy_errors_total 2
```

## Benchmark

キャッシュの有無によるレイテンシ差を測定する予定です。

| Condition | p50 | p95 | Notes |
|---|---:|---:|---|
| First request | TBD | TBD | origin fetch + resize + encode |
| Cache hit | TBD | TBD | disk cache |
| WebP conversion | TBD | TBD | encode cost included |

測定には `hey` などのHTTP負荷試験ツールを利用します。

```bash
hey -n 100 -c 10 "http://localhost:8080/image?url=https://example.com/sample.jpg&w=512&format=jpeg"
```

## Getting Started

### Requirements

- Go 1.22+
- Git

### Run

```bash
git clone https://github.com/yourname/mini-imageflux-go.git
cd mini-imageflux-go
go run ./cmd/server
```

### Example Request

```bash
curl "http://localhost:8080/image?url=https://example.com/sample.jpg&w=512&format=jpeg" -o output.jpg
```

## Tech Stack

- Go
- `net/http`
- `image/jpeg`
- `image/png`
- WebP encoder
- Disk cache
- Prometheus format metrics
- nginx

## Directory Structure

```txt
mini-imageflux-go/
  cmd/
    server/
      main.go
  internal/
    cache/
    fetcher/
    imageproc/
    metrics/
    server/
  docs/
    design.md
    benchmark.md
  README.md
  go.mod
```

## Current Status

このプロジェクトは現在開発初期段階です。

短期的には、以下のMVPを実装します。

- URLから画像を取得する
- 横幅指定でリサイズする
- JPEG / WebPで返す
- 変換済み画像をディスクキャッシュする
- `/metrics` で基本的なメトリクスを返す
- READMEに設計・計測結果を記録する

## Known Issues

現時点で想定している課題は以下です。

- オリジンURLの検証が不十分
- SSRF対策が必要
- 巨大画像によるメモリ使用量増大への対策が必要
- キャッシュ削除ポリシーが未実装
- 同一画像への同時リクエスト時に重複変換が発生する可能性がある
- WebP / AVIF変換時のCPU負荷をまだ測定できていない
- オリジン取得失敗時のリトライ・タイムアウト設計が未整理
- キャッシュ破損時の扱いが未実装

## Future Work

今後は以下の機能を追加する予定です。

- nginx reverse proxy構成
- nginx proxy cacheの検証
- URL署名
- 期限付きURL
- AVIF対応
- Range Request対応
- worker poolによる変換処理制御
- rate limit
- pprofによるCPU / memory profiling
- Prometheus / Grafanaによる可視化
- TypeScript / Vueによる簡易管理画面

## Notes

このプロジェクトは、画像変換プロキシを題材に、画像配信基盤の基本要素を検証するための実験的な実装です。  
本番運用を目的としたものではありません。