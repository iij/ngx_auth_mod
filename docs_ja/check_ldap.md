# check\_ldap

[ngx\_ldap\_auth](ngx_ldap_auth.md)または[ngx\_ldap\_path\_auth](ngx_ldap_path_auth.md)の設定ファイルを使って、LDAPサーバで認証処理をチェックするためのプログラムです。
動作確認用です。

## エラー処理

エラー時は、プロセスが異常ステータスで終了します。 

## 実行方法

コマンドラインは、以下の通りです。

```
check_ldap 設定ファイル名 アカウント名
```

設定ファイルとユーザ名を引数にして、実行します。
実行後、パスワードを入力すると、結果を出力します。

設定ファイルは、[ngx\_ldap\_auth](ngx_ldap_auth.md)と同じものです。
