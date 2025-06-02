# Hexabase KaaS: アーキテクチャー仕様書

## 1. プロジェクト概要

### 1.1. ビジョン

本プロジェクトは、`K3s`と`vCluster`を基盤とした、監視可能で直感的なマルチテナント Kubernetes as a Service (KaaS) プラットフォームを、オープンソースで開発・提供することを目的とします。現代のアプリケーション開発において Kubernetes はデファクトスタンダードとなりつつありますが、その学習コストの高さ、運用管理の複雑さ、そして小規模チームや個人開発者にとってはリソース確保の難しさが導入の障壁となっています。Hexabase KaaS は、これらの課題を解決するために設計されました。

具体的には、以下の価値を提供します。

- **導入の容易性**: `K3s`という軽量な Kubernetes ディストリビューションをベースとし、さらに`vCluster`による仮想化技術を用いることで、ユーザーは物理的なクラスタ管理の複雑さから解放され、迅速に隔離された Kubernetes 環境を利用開始できます。この容易性は、特に Kubernetes の専門知識が限られている開発者やチームにとって、新しい技術への挑戦を後押しします。従来の Kubernetes クラスタ構築には、ネットワーク設定、セキュリティポリシーの策定、ストレージのプロビジョニングなど、多岐にわたる専門知識が必要でしたが、Hexabase KaaS ではこれらの多くを自動化・抽象化し、数クリックで利用可能な状態を提供します。

- **直感的な操作性**: Kubernetes の強力な機能を、専門知識がないユーザーでも容易に扱えるよう、洗練された UI/UX を通じて抽象化します。Organization、Workspace、Project といった直感的な概念でリソースを管理できます。例えば、アプリケーションのデプロイメント、スケーリング、モニタリングといった一般的な操作を、YAML ファイルを直接編集することなく、グラフィカルなインターフェースから実行できるようにします。また、エラーメッセージやログも分かりやすく表示し、問題解決を支援します。

- **強力なテナント分離**: `vCluster`は、各テナント（Workspace）に専用の API サーバーとコントロールプレーンコンポーネントを提供することで、Namespace ベースの分離よりも格段に高いセキュリティと独立性を保証します。これにより、テナント間の影響を最小限に抑え、安心してリソースを利用できます。例えば、あるテナントが誤ってリソースを大量に消費したり、セキュリティ上の問題を引き起こしたりしても、他のテナントの環境には影響が及ばないように設計されています。これは、特に複数の顧客やプロジェクトを同一の物理インフラ上でホストする場合に極めて重要です。

- **クラウドネイティブな運用**: `Prometheus`、`Grafana`、`Loki`による包括的な監視スタックを標準装備し、システムの健全性やテナントのリソース使用状況をリアルタイムに把握できます。`Flux`による GitOps アプローチを採用することで、宣言的な構成管理と再現性の高いデプロイメントを実現します。`Kyverno`によるポリシー管理機能は、セキュリティコンプライアンスとガバナンスの強化を支援します。これにより、インフラストラクチャの構成変更、アプリケーションのデプロイ、セキュリティポリシーの適用などが、バージョン管理されたコードを通じて行われ、監査証跡の確保やロールバックが容易になります。

- **オープンソースとしての透明性とコミュニティ**: 本プロジェクトをオープンソースとして公開することで、技術的な透明性を確保し、世界中の開発者からのフィードバックやコントリビューションを積極的に受け入れます。コミュニティと共に成長し、より多くのユースケースに対応できる、信頼性の高いプラットフォームを構築することを目指します。教育機関での利用や、新しいクラウドネイティブ技術の学習・検証プラットフォームとしての活用も期待しています。ソースコードの公開は、セキュリティの脆弱性を早期に発見し、修正することにも繋がります。また、多様な視点からの意見を取り入れることで、より革新的で実用的な機能開発が可能になります。

Hexabase KaaS は、Kubernetes のパワーをより多くの人々に届け、イノベーションを加速させるための触媒となることをビジョンとして掲げています。開発者はインフラの複雑さから解放され、アプリケーション開発という本質的な価値創造に集中できるようになります。

### 1.2. 既存コードベース

本プロジェクトは、以下の既存コードベースを基盤としつつ、本仕様書に基づき大幅な再実装および機能拡張を行います。これらのリポジトリは、プロジェクトの初期段階におけるアイデアやプロトタイピングの成果物であり、本格的な KaaS プラットフォームとしての機能と品質を実現するためには、抜本的な見直しと再構築が必要です。

- **UI (Next.js)**: <https://github.com/b-eee/hxb-next-webui>

  - **現状の評価**: 現状のリポジトリには、基本的なダッシュボードのレイアウト構想、UI コンポーネントライブラリの選定（例: Material UI, Ant Design など）、いくつかの画面の静的なモックアップ、または初期のプロトタイプが含まれている可能性があります。しかし、KaaS としての複雑な状態管理、リアルタイムなデータ表示、ユーザーロールに応じた動的な UI 制御などの機能は未実装であると想定されます。
  - **再実装・拡張方針**: 本仕様書で定義される Hexabase KaaS のコアコンセプト（Organization、Workspace、Project、Role、Group など）に対応するため、UI 全体の情報アーキテクチャから再設計し、コンポーネントベースの開発を徹底します。状態管理には、グローバルな状態とローカルな状態を効率的に管理できるライブラリ（例: Zustand, Recoil, Redux Toolkit）を導入します。API サーバーとのデータ通信には、キャッシュ機能や自動再フェッチ機能を提供するデータフェッチライブラリ（例: SWR, React Query）を採用し、リアルタイム性の高いユーザー体験を目指します。特に、vCluster のプロビジョニング状況の表示、リソース使用量のグラフ表示、ログのストリーミング表示、ユーザーへの通知機能（例: WebSocket や Server-Sent Events を利用）などが重要な開発項目となります。API との連携部分は、本仕様書で定義される API エンドポイントの仕様に基づいて、型安全なクライアントコードを生成する（例: OpenAPI Generator）など、開発効率と品質を高めるアプローチを検討します。アクセシビリティ（a11y）や国際化（i18n）も初期段階から考慮に入れます。

- **API (Go)**: <https://github.com/b-eee/apicore>
  - **現状の評価**: 既存の API コアは、汎用的な REST API の基盤、もしくは特定の機能に特化した初期プロトタイプである可能性があります。KaaS コントロールプレーンとしての複雑なビジネスロジック、テナント管理、Kubernetes リソース操作、外部サービス連携（OIDC, Stripe など）といった機能は、大幅な追加開発または再設計が必要となるでしょう。
  - **再実装・拡張方針**: 本仕様書で定義されるコントロールプレーンの全機能を Go 言語で再実装します。これには、堅牢な RESTful API サーバーの構築（例: Gin, Echo フレームワークの利用）、OIDC プロバイダー機能の実装、`client-go`ライブラリを駆使した vCluster オーケストレーションロジック、Stripe SDK を利用した課金システム連携、NATS クライアントを利用した非同期処理ワーカーの開発が含まれます。データベーススキーマ（PostgreSQL）は本仕様書に基づいて新規に設計し、GORM などの ORM を利用して効率的なデータアクセスを実現します。エラーハンドリング、構造化ロギング（例: Logrus, Zap）、メトリクス収集（Prometheus クライアントライブラリ）、設定管理（Viper など）といった運用上重要な要素も初期から組み込みます。既存コードで再利用可能なユーティリティ関数や基本的な構造（例: HTTP サーバーのセットアップ、基本的なミドルウェア）があれば活用しますが、KaaS 特有のドメインロジックやワークフローは、本仕様書に基づいて新規に、かつテスト駆動開発（TDD）やドメイン駆動設計（DDD）の原則を意識しながら開発を進めます。

## 2. システムアーキテクチャ

Hexabase KaaS のシステムアーキテクチャは、ユーザーが直接操作する**Hexabase UI (Next.js)**、システム全体の管理と制御を行う**Hexabase API (コントロールプレーン、Go 言語)**、そしてこれらを支える各種**ミドルウェア（PostgreSQL, Redis, NATS など）**から構成されます。これらのコンポーネントは全てコンテナ化され、運用基盤となる**Host K3s Cluster**上で稼働します。テナントごとの Kubernetes 環境は、**vCluster**技術によって Host K3s Cluster 内に仮想的に構築され、強力な分離と独立性を提供します。この多層的なアーキテクチャは、スケーラビリティ、可用性、保守性を考慮して設計されています。

**主要コンポーネント間の連携とデータの流れ:**

1.  **ユーザー操作と UI**: ユーザーは Web ブラウザを通じて Hexabase UI にアクセスし、Organization の作成、Workspace（vCluster）のプロビジョニング、Project（Namespace）の管理、ユーザー招待、権限設定などの操作を行います。UI はこれらの操作を Hexabase API へのリクエストに変換します。UI はユーザーの認証状態を管理し、API リクエストには認証トークンを付与します。リアルタイムな情報更新（例: vCluster のプロビジョニング進捗）は、WebSocket や Server-Sent Events などの技術を用いて実現することを検討します。

2.  **API リクエスト処理**: Hexabase API は、UI からのリクエストを受け付け、まず認証・認可処理を行います。認証されたユーザーが要求された操作を実行する権限を持つかを確認した後、ビジネスロジックを実行します。これには、PostgreSQL データベースの状態を更新したり、vCluster オーケストレーターに指示を出したりする処理が含まれます。時間のかかる処理（例: vCluster 作成、大規模な設定変更）は、API サーバーの応答性を損なわないよう、NATS メッセージキューにタスクとして登録し、非同期ワーカーに処理を委譲します。API はリクエストのバリデーションも厳格に行い、不正な入力に対しては適切なエラーレスポンスを返します。

3.  **vCluster オーケストレーション**: vCluster オーケストレーターは、Host K3s クラスタと対話し、vCluster のライフサイクル（作成、設定、削除）を管理します。具体的には、`vcluster CLI`や Kubernetes API (`client-go`) を用いて、vCluster の Pod（実体は StatefulSet や Deployment）をデプロイし、必要なネットワーク設定（Service、Ingress など）やストレージ設定（PersistentVolumeClaim）を行います。また、各 vCluster に対する OIDC 設定の適用、HNC (Hierarchical Namespace Controller) のインストールと設定、テナントプランに応じたリソースクオータの設定、Dedicated Node の割り当て制御（Node Selector や Taints/Tolerations の利用）なども担当します。さらに、vCluster 内の Namespace や RBAC（Role, RoleBinding, ClusterRole, ClusterRoleBinding）の設定も、ユーザーの操作に応じてこのコンポーネントが実行します。

4.  **非同期処理**: 非同期ワーカーは、NATS メッセージキューからタスクを受け取り、vCluster のプロビジョニング、Stripe API との連携（課金処理）、HNC セットアップ、バックアップ・リストア処理（将来的な機能）などのバックグラウンド処理を実行します。これにより、API サーバーは長時間ブロックされることなく、迅速にレスポンスを返すことができます。ワーカーは処理の進捗状況をデータベースに記録し、完了またはエラー発生時には NATS を通じて API サーバーや通知システムに結果を通知することを検討します。

5.  **状態永続化**: PostgreSQL データベースは、Organization、Workspace、Project、User、Group、Role、課金プラン、サブスクリプション情報、非同期タスクの状況、監査ログなどが格納されます。データの整合性を保つためにトランザクションを適切に使用し、定期的なバックアップとリストア戦略も計画します。スキーマの変更はマイグレーションツール（例: golang-migrate）を用いて管理します。

6.  **キャッシュ**: Redis は、ユーザーセッション情報、頻繁にアクセスされる設定データ、OIDC トークンの検証に必要な公開鍵（JWKS）、レートリミットのカウンターなどをキャッシュし、データベースへの負荷を軽減し、システムの応答性とスケーラビリティを向上させます。キャッシュの有効期限や無効化戦略も適切に設計します。

7.  **監視とロギング**: Prometheus はシステム全体のメトリクス（API サーバーのパフォーマンス指標、vCluster のリソース使用状況、NATS のキュー長、PostgreSQL の接続数など）を収集します。Loki は全てのコンポーネント（API サーバー、ワーカー、vCluster のコントロールプレーンログなど）のログを一元的に集約します。Grafana はこれらのデータを可視化し、運用者がシステムの健全性をリアルタイムに監視し、問題発生時に迅速に対応するためのダッシュボードを提供します。アラートは Alertmanager を通じて運用チームに通知されます。

8.  **GitOps によるデプロイ**: Hexabase コントロールプレーン自体のデプロイや更新は、Flux を用いた GitOps ワークフローによって管理されます。インフラストラクチャ構成（Kubernetes マニフェスト、Helm Chart）、アプリケーション設定、セキュリティポリシーなどは全て Git リポジトリで宣言的に管理されます。変更は Git へのコミットとプルリクエストを通じて行われ、承認されると Flux が自動的に Host K3s クラスタに適用します。これにより、デプロイメントの再現性、監査性、信頼性が向上します。

9.  **ポリシー適用**: Kyverno は、Kubernetes Admission Controller として動作し、Host K3s クラスタおよび各 vCluster 内（設定可能であれば）で、セキュリティポリシーや運用ポリシーを強制します。例えば、「全ての Namespace には`owner`ラベルを必須とする」「特権コンテナの起動を禁止する」「信頼できないイメージレジストリからのイメージプルをブロックする」といったポリシーを定義し、コンプライアンスを維持します。ポリシーも GitOps を通じて管理されることが望ましいです。

このアーキテクチャにより、スケーラブルで、回復力があり、運用しやすい KaaS プラットフォームの実現を目指します。各コンポーネントの役割分担を明確にし、標準化された技術とオープンソースプロダクトを活用することで、開発効率とシステムの信頼性を高めます。

## 3. コアコンセプトとエンティティマッピング

Hexabase KaaS は、ユーザーが Kubernetes の複雑さを意識せずにサービスを利用できるよう、独自の抽象化された概念を提供します。これらの概念は、内部的には Kubernetes の標準的なリソースや機能にマッピングされます。このマッピングを理解することは、システムの動作を把握し、効果的に利用する上で非常に重要です。

| Hexabase 概念         | Kubernetes 相当           | スコープ           | 備考                                                      |
| --------------------- | ------------------------- | ------------------ | --------------------------------------------------------- |
| Organization          | (なし)                    | Hexabase           | 課金・請求、組織ユーザー管理の単位。ビジネスロジック。    |
| Workspace             | vCluster                  | Host K3s Cluster   | 強力なテナント分離境界。                                  |
| Workspace Plan        | ResourceQuota / Node 構成 | vCluster / Host    | リソース上限を定義。                                      |
| Organization User     | (なし)                    | Hexabase           | 組織の管理・課金担当者。                                  |
| Workspace Member      | User (OIDC Subject)       | vCluster           | vCluster を操作する技術担当者。OIDC で認証される。        |
| Workspace Group       | Group (OIDC Claim)        | vCluster           | 権限付与の単位。階層は Hexabase が解決する。              |
| Workspace ClusterRole | ClusterRole               | vCluster           | Workspace 全体に及ぶプリセット権限（例: Admin, Viewer）。 |
| Project               | Namespace                 | vCluster           | Workspace 内のリソース分離単位。                          |
| Project Role          | Role                      | vCluster Namespace | Project 内でユーザーがカスタム作成可能な権限。            |

# 4. 機能仕様とユーザーフロー

## 4.1. サインアップと組織管理

- **新規ユーザー登録**  
  外部 IdP（Google, GitHub 等）による OpenID Connect でサインアップ。Hexabase DB に User が作成される。

- **Organization 作成**  
  初回サインアップ時、ユーザーのプライベートな Organization が自動作成される。ユーザーはこの Org の最初の Organization User となる。

- **Organization 管理**  
  Organization User は、課金情報（Stripe 連携）の管理や、他の Organization User の招待を行える。  
  ※この権限では、配下の Workspace（vCluster）内のリソースを直接操作することはできない。

## 4.2. Workspace (vCluster) の管理

- **作成**  
  Organization User は、Plan（リソース上限）を選択して新しい Workspace を作成する。

- **プロビジョニング**  
  Hexabase コントロールプレーンが Host クラスタ上に vCluster をプロビジョニングし、自身を信頼する OIDC プロバイダーとして設定する。

- **初期設定 (vCluster 内)**

  - プリセット ClusterRole の作成:  
    `hexabase:workspace-admin` と `hexabase:workspace-viewer` の 2 つの ClusterRole を自動作成。  
    ※ユーザーによるカスタム ClusterRole の作成は禁止。
  - デフォルト ClusterRoleBinding の作成:  
    `hexabase:workspace-admin` ClusterRole を `WSAdmins` グループに紐付ける ClusterRoleBinding を自動作成。

- **初期設定 (Hexabase DB 内)**
  - デフォルトグループ作成:  
    `WorkspaceMembers`（最上位）, `WSAdmins`, `WSUsers` の 3 つを階層構造で作成。
  - Workspace 作成者を `WSAdmins` グループに所属させることで、vCluster の管理者となる。

## 4.3. Project (Namespace) の管理

- **作成**  
  Workspace Member（WSAdmins 等、権限を持つユーザー）が Workspace 内に新しい Project を作成。

- **Namespace 作成**  
  Hexabase コントロールプレーンが vCluster 内に対応する Namespace を作成。

- **ResourceQuota 適用**  
  Workspace の Plan で定義されたデフォルトの ResourceQuota オブジェクトを Namespace に自動作成。

- **カスタム Role 作成**  
  Project (Namespace) 内で有効なカスタム Role を UI から作成・編集可能。

## 4.4. 権限管理と継承

- **権限割り当て**  
  UI を通じて、Project Role やプリセット ClusterRole を Workspace Group に割り当て。  
  Hexabase は RoleBinding や ClusterRoleBinding を vCluster 内に作成・削除。

- **権限継承の解決**
  - ユーザーが vCluster にアクセスする際、OIDC プロバイダーは以下を実行:
    1. 所属グループとその親グループを DB から再帰的に取得。
    2. フラットなグループリストを OIDC トークンの `groups` クレームに含めて発行。
    3. vCluster の API サーバーがこの情報に基づいてネイティブに RBAC 認可を実行。

---

# 5. 技術スタックとインフラストラクチャ

## 5.1. アプリケーション

- **フロントエンド**: Next.js
- **バックエンド**: Go (Golang)

## 5.2. データストア

- **プライマリ DB**: PostgreSQL
- **キャッシュ**: Redis

## 5.3. メッセージングと非同期処理

- **メッセージキュー**: NATS

## 5.4. CI/CD（継続的インテグレーション/デリバリー）

- **パイプラインエンジン**: Tekton

  - **理由**: Kubernetes ネイティブで宣言的なパイプラインを構築可能。コンテナビルド、テスト、セキュリティスキャンを自動化。

- **デプロイメント（GitOps）**: ArgoCD または Flux
  - **理由**:  
    Git リポジトリを信頼できる唯一の情報源（Source of Truth）とし、クラスタの状態を宣言的に管理。  
    ArgoCD は UI が強力で、Flux はシンプルさと拡張性に優れる。プロジェクトの好みに応じて選択可能。

## 5.5. セキュリティとポリシー管理

- **コンテナ脆弱性スキャン**: Trivy

  - **役割**:  
    CI パイプライン（Tekton）に組み込み、コンテナイメージのビルド時に OS パッケージや言語ライブラリの既知の脆弱性（CVE）をスキャン。IaC の構成ミスも検出可能。

- **ランタイムセキュリティ監査**: Falco

  - **役割**:  
    ランタイム脅威検知ツール（CNCF 卒業プロジェクト）。カーネルレベルでシステムコールを監視し、「コンテナ内での予期せぬシェルの起動」や「機密ファイルへのアクセス」などをリアルタイムで検知・アラート。

- **ポリシー管理エンジン**: Kyverno
  - **Kyverno**:  
    Kubernetes リソース（YAML）としてポリシーを記述できるため学習コストが低く、「特定のラベルがない Pod の作成禁止」「信頼できないイメージレジストリの使用ブロック」などを直感的に管理可能。

---

# 6. インストールとデプロイ（IaC）

本プロジェクトでは「簡単なインストール」を実現するために、IaC（Infrastructure as Code）として **Helm** を採用。

## 6.1. Helm Umbrella Chart

Hexabase の全コンポーネントと依存ミドルウェアを、単一コマンドでデプロイ可能な Helm Umbrella Chart として提供。

```yaml
apiVersion: v2
name: hexabase-kaas
description: A Helm chart for deploying the Hexabase KaaS Control Plane
version: 0.1.0
appVersion: "0.1.0"

dependencies:
  # 依存する公式・コミュニティのHelm Chartを定義
  - name: postgresql
    version: "14.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled # 必要に応じて無効化可能
  - name: redis
    version: "18.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: redis.enabled
  - name: nats
    version: "1.x.x"
    repository: "https://nats-io.github.io/k8s/helm/charts/"
    condition: nats.enabled
```

**Chart 内のテンプレート例（`templates/`）**:

- Hexabase API（Go）の Deployment / Service
- Hexabase UI（Next.js）の Deployment / Service
- DB 接続情報などの Secret（初回インストール時に自動生成）
- 各種設定を管理する ConfigMap

## 6.2. インストールフロー

エンドユーザーは、K3s クラスタを準備後、以下の手順で Hexabase KaaS をデプロイ可能：

### Helm リポジトリの追加

```bash
helm repo add hexabase https://<your-chart-repository-url>
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### 設定ファイル (values.yaml) の編集 (オプション):

- ドメイン名やリソース割り当てなど、カスタマイズが必要な項目を編集。

### Helm によるインストール:

```bash
helm install hexabase-kaas hexabase/hexabase-kaas -f values.yaml
```

この単一のコマンドにより、PostgreSQL、Redis、NATS といった依存コンポーネントと共に、Hexabase コントロールプレーン全体が K3s クラスタ上にセットアップされます。

# 7. 結論

本仕様書は、モダンな技術スタックとクラウドネイティブのベストプラクティスに基づいた、Hexabase KaaS のコンセプト設計図です。Helm によるシンプルな導入、Tekton と GitOps による効率的な CI/CD、Trivy と Falco による堅牢なセキュリティ、そして Kyverno による柔軟なポリシー管理を取り入れることで、世界中のユーザーに安心して利用され、コミュニティと共に成長していけるオープンソースプロジェクトの強固な基盤を築きます。
