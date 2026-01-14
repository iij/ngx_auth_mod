# 認証パラメータの書き換え

以下のモジュールは、書き換えたHTTPリクエストの値で認証を呼び出します。

- [ngx\_rewrite\_auth](ngx_rewrite_auth.md)
- [ngx\_rewrite\_and\_auth](ngx_rewrite_and_auth.md)
- [ngx\_rewrite\_switch\_auth](ngx_rewrite_switch_auth.md)

このドキュメントでは、この認証パラメータの書き換えに使う書式について説明します。

## 認証パラメータの書式

認証パラメータは、以下の書式で記述します。

| 書式 | 変換後の文字列 |
| :--- | :--- |
| `$$` | `$` |
| `$u` | Basic認証のユーザ名 |
| `$p` | Basic認証のパスワード |
| `$h{ヘッダー名}` | 指定されたヘッダー名のHTTPヘッダーの文字列 |
| `$0` | **username\_re** の正規表現のパターン全体にマッチした文字列 |
| `$1`〜`$9` | **username\_re** の正規表現グループにマッチした文字列(正規表現グループの1〜9番目に対応) |
| 上記以外 | 記述された文字列 |

## 設定できる認証パラメータ

設定できる認証パラメータは、以下の通りです。

| HTTPリクエストの要素 | 設定ファイルのパラメータ名 | 補足 |
| :--- | :--- | :--- |
| Basic認証ユーザー名 | **set\_usename** | 元のユーザ名を設定したい時は、**set\_usename**を`$u`に設定してください。デフォルトでは空になります。 |
| Basic認証パスワード | **set\_password** | 元のパスワードを設定したい時は、**set\_password**を`$p`に設定してください。デフォルトでは空になります。|
| HTTPヘッダー | **\[\[set\_header\]\]**<br>**\[\[and.set\_header\]\]**<br>**\[\[swtich.set\_header\]\]** | デフォルトで設定されるヘッダーは、`User-Agent`および`Cache-Control`だけです。 |

## 変換例
以下は、変換例です。　

| **username\_re**の正規表現 | 入力値 | 設定書式 | 変換後の値 |
| :--- | :--- | :--- | :--- |
| 空文字列 | ユーザー名が`user1` |`$u@example.com` | `user1@example.com` |
| 空文字列 | パスワードが`secret` |`$p` | `secret` |
| `^([^@]+)@example.com$` | ユーザー名が`user2@example.com` | `$1` | `user2` |
| `^[^@]+@(.+)$` | ユーザー名が`user3@example.com` | `$1` | `example.com` |
| 空文字列 | `X-Authz-Path`ヘッダーの値が`markdown` |`/$h{X-Authz-Path}/` | `/markdown/` |
