# ict-school-web


このプロジェクトは「模擬授業予約システム」のための Go ウェブアプリケーションです。**Nix** を使用して再現性のあるビルドを行い、**Docker** を使用してデプロイします。

## 0. セットアップとインストール

プロジェクトを実行する前に、必要なツールをインストールしてください。

### A. ローカルマシン（ビルド作業用）
必要なのは **Nix** だけです。Go、Postgres、Docker をローカルにインストールする必要はありません（Nix が管理します）。

**1. Nix のインストール:**
Mac / Linux の場合:
```bash
sh <(curl -L [https://nixos.org/nix/install](https://nixos.org/nix/install)) --daemon

```

**2. Flakes の有効化（必須）:**
Nix Flakes は実験的機能であるため、設定で有効にする必要があります。

```bash
mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf

```

*インストール後はターミナルを再起動してください。*

---

### B. サーバー（実行環境用）

AWS EC2（Amazon Linux 2023 等）などのサーバー側では、ビルドしたイメージを実行するために **Docker** と **Docker Compose** が必要です。

**1. Docker のインストール:**

```bash
sudo yum update -y
sudo yum install -y docker
sudo service docker start
sudo usermod -a -G docker ec2-user

```

*（権限設定を反映させるため、一度ログアウトして再ログインしてください）*

**2. Docker Compose のインストール:**

```bash
mkdir -p /usr/libexec/docker/cli-plugins/
curl -SL [https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64](https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64) -o /usr/libexec/docker/cli-plugins/docker-compose
chmod +x /usr/libexec/docker/cli-plugins/docker-compose

```

**3. インストールの確認:**

```bash
docker compose version
# 出力例: Docker Compose version v2.x.x

```

---

## 1. ローカル開発 (Local Development)

コードを修正したり、ローカルでテストしたりする手順です。

1. **開発シェルに入る:**
このコマンドで Go や Postgres のダウンロード、環境変数の設定が自動的に行われます。
```bash
nix develop

```


2. **サーバーを起動する:**
```bash
go run ./cmd/server/main.go

```


サーバーは `http://localhost:8080` で起動します。

---

## 2. 本番用ビルド (Building for Production)

Nix を使用して、どの環境でも完全に同一（byte-for-byte）な Docker イメージを作成します。

1. **イメージのビルド:**
```bash
nix build .#docker

```


このコマンドが完了すると、フォルダ内に `result` という名前のシンボリックリンクが作成されます。これが Docker イメージの圧縮アーカイブです。

---

## 3. デプロイ手順 (Deployment)

### ステップ 1: ファイルをサーバーへ転送

AWS の秘密鍵 (`key.pem`) とサーバーの IP アドレス (`1.2.3.4`) があると仮定します。

```bash
# 1. ビルドした Docker イメージ（圧縮ファイル）を転送
scp -i key.pem result ec2-user@1.2.3.4:/home/ec2-user/ict-web.tar.gz

# 2. 設定ファイルを転送
scp -i key.pem docker-compose.yml init.sql ec2-user@1.2.3.4:/home/ec2-user/

```

### ステップ 2: サーバーでの読み込みと実行

サーバーに SSH でログインします:

```bash
ssh -i key.pem ec2-user@1.2.3.4

```

**サーバー内での操作:**

```bash
# 1. イメージを Docker に読み込む
docker load < ict-web.tar.gz

# 2. アプリケーションを起動する
# (Postgres の起動、init.sql の実行、Go アプリの起動が一括で行われます)
docker compose up -d

```

### ステップ 3: ステータス確認

```bash
docker compose ps

```

`db` と `web` の両方の Status が `Up` になっていれば成功です。

---

## トラブルシューティング

**Q: SSH 接続時に "Permission denied" と出る**
秘密鍵ファイルの権限が適切でない可能性があります。自分だけが読める設定にしてください。

```bash
chmod 400 key.pem

```

**Q: ログに "Database connection refused" と出る**
`docker-compose.yml` を確認してください。`DATABASE_URL` は `localhost` ではなく、サービス名である `db` を指している必要があります。

```yaml
DATABASE_URL: postgres://postgres:pass@db:5432/ict?sslmode=disable

```

**Q: コードを更新したので再デプロイしたい**

1. ローカルで `nix build .#docker` を実行して再ビルドします。
2. 新しくできた `result` を `scp` でサーバーに転送します。
3. サーバー側で `docker load < ict-web.tar.gz` を実行します。
4. サーバー側で `docker compose up -d` を実行します（Docker が新しいイメージを検知して、Web アプリだけを再起動します）。

```

```
