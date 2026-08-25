# 1. Goのバージョンを指定
FROM golang:1.26.4

# 2. 必要なパッケージをインストール
RUN apt-get update && apt-get install -y git

# 3. 作業ディレクトリを設定
WORKDIR /go/src/app

# 4. ホストマシンからアプリケーションファイルをコンテナにコピー
COPY ./app .

# 5. Goアプリケーションを実行
CMD ["go", "run", "."]
