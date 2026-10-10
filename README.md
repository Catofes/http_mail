## Go / Echo

邮箱管理 API，使用 Go、Echo v4 和 PostgreSQL。默认监听 `0.0.0.0:8000`，直接使用旧程序的数据库表，不自动建表或修改表结构。

### 运行

需要 Go 1.26.1 或更新版本编译；编译后的程序不依赖 Python 或系统 `libcrypt`。

```sh
cp config.json.example config.json
# 编辑 config.json，填写现有数据库的连接信息和原来的 password_salt。
chmod 600 config.json
make build
./http_mail
```

可以使用 `./http_mail -config /path/to/config.json -listen 0.0.0.0:8000` 指定配置和监听地址。

从原来的 `config.py` 迁移配置时，把 `password_salt`、`db_user`、`db_passwd`、`db_host`、`db_port`、`db_db`、`db_mincached`、`db_maxconnections` 和 `session_cache_size` 原值复制到 JSON。**不要改变 `password_salt`，否则原有管理账户无法登录。** 示例端口为 PostgreSQL 常用的 5432，实际使用时保留原来的连接端口。原代码中的 `db_maxcached`、`db_maxshared` 没有被数据库实现使用，不需要迁移。

`db_sslmode` 默认 `prefer`，与原来 psycopg2 的默认行为一致；也可以按数据库配置填写 `disable`、`require`、`verify-ca` 或 `verify-full`。数据库密码只保存在配置文件中，不打包进容器镜像。

### 兼容范围

- 保留原有 17 个路径的 HTTP 方法、查询参数、JSON 字段、HTTP 状态和业务错误码。错误 JSON 仍为 `{"title":"…","code":…}`，不采用 Echo 默认的 `message` 格式。
- 写入成功仍返回 HTTP 200 和空正文；空列表仍返回 `{"result":null}`；没有 DKIM 记录仍返回空正文。
- 管理账户密码仍为 `sha512(password + password_salt)` 的十六进制结果前 64 个字符；邮箱密码仍为 `$6$` SHA512-crypt，默认 5000 轮、16 字符盐，保持 Postfix / Dovecot 使用的密码格式。
- token 仍通过 URL 查询参数传递。保留原代码的十六进制 token 格式（最多 16 字符），内存缓存和固定 24 小时有效期。重启服务后需要重新登录；修改密码不会主动清除现有会话。
- 保留域名所有权检查，以及只有 `level == 100` 可以越过所有权检查的规则。Alias、BCC、Transport 写入要求 `level >= 5`；DKIM、邮箱用户和默认 Transport 保留原来的权限规则。
- `/server` 和默认路由操作均读取生产表 `my_networks`。修正旧代码中默认路由查询 `mynetworks` 的拼写错误，无需修改数据库或创建兼容视图。
- `/transport_default/{domain_id}/1` 保留实际代码行为：检查服务器 ID 后，事务内清空该域名的 Transport，并写入 `lmtp:unix:private/dovecot-lmtp` / `0default`。下面的历史 API 说明文字保留，操作行为以本条为准。

旧 Python 源码和原 README 保存在 `legacy/`，便于核对和回退，不参与 Go 构建。启动方式与配置文件格式已改变，邮箱数据库无需数据转换。未新增 API、认证机制或邮件功能。

### 验证

```sh
make check        # 格式、vet、单元测试和 race 检查
make integration  # 临时 PostgreSQL 数据库，验证全部接口与数据库写入
```

集成测试需要安装 PostgreSQL 的 `initdb`、`pg_ctl`，以普通用户运行。它只在 `/tmp` 创建临时数据库，关闭 TCP 监听，测试结束后清理。也可以显式设置 `HTTP_MAIL_TEST_DSN` 指向专用测试数据库；测试只创建、清理自己的随机命名 schema。`testdata/schema.sql` 是根据旧代码涉及的字段构造的测试夹具，不是生产数据库迁移文件。

### Docker

```sh
make docker-build docker-check IMAGE=http_mail:local
docker run --rm -p 8000:8000 \
  -v "$PWD/config.json:/app/config.json:ro" http_mail:local
```

镜像使用多阶段构建，运行层只有静态 Go 程序与 CA 证书。容器内的 `db_host` 应填写数据库的可达地址。

### GitHub Actions 与 tag 发布

[Build And Release](.github/workflows/build-release.yml) 仿照 Attendance 的发布流程，使用 Makefile 执行构建与检查：

- PR、推送到 `main` / `master` 或手动运行：执行 `make check`、独立 PostgreSQL 接口测试、`make build release`，并在 amd64、arm64 原生 runner 上构建、检查 Docker 镜像。普通提交只验证构建，不发布镜像或 Release。
- 推送 `v*` tag：上述检查通过后，将两个架构的镜像推送到 GHCR，合并并验证多架构清单，再创建 GitHub Release。
- GHCR 地址为 `ghcr.io/catofes/http_mail`（fork 使用自己的小写仓库路径）。镜像标签包含版本 tag 与提交 SHA；正式版本更新 `latest`。带 `-` 的 tag，例如 `v0.1.0-rc.1`，标记为预发布，不更新 `latest`。
- Release 附件包含 `http_mail-linux-amd64`、`http_mail-linux-arm64`、`config.json.example`、`README.md` 和 `SHA256SUMS`。构建后以及 Release 上传前都会验证校验和。

在本地生成相同的发布附件：

```sh
make release
cd dist
sha256sum --check SHA256SUMS
```

发布时，在已检查并提交的版本上创建、推送 tag，例如：

```sh
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

可以在 `docs/releases/<tag>.md` 提供版本说明；没有该文件时，工作流自动生成说明。发布使用 GitHub 内置的 `GITHUB_TOKEN`，仓库需允许 Actions 创建 Release 和写入 GHCR packages。推送 tag 后，应分别确认 Actions 成功、Release 附件齐全以及 GHCR 的 amd64 / arm64 清单。此流程不部署邮箱服务，也不会携带真实数据库配置。

## API

### Common

The Request must encode in JSON format. And all the api will return JSON if there is something needed to be returned.

#### Error Code:

If everything worked perfect, a HTTP 200 will returned. All other HTTP STATUS CODE means some errors. 
A error code and a illustrate will be returned. Error codes are shown below.

    Error_text = {
        0: ['Unknown Error.',
            falcon.HTTP_500],
        1: ['SQL Error.',
            falcon.HTTP_500],
        2: ['Json Required.',
            falcon.HTTP_400],
        3: ['Empty Request Body.',
            falcon.HTTP_400],
        4: ['Malformed JSON.',
            falcon.HTTP_400],
        5: ['Required Invite Code.',
            falcon.HTTP_400],
        6: ['Error Invite Code.',
            falcon.HTTP_404],
        7: ['Domain Required.',
            falcon.HTTP_400],
        8: ['Domain Duplicated.',
            falcon.HTTP_403],
        9: ['Domain Illegal.',
            falcon.HTTP_403],
        10: ['Domain Id Required.',
             falcon.HTTP_400],
        11: ['Error Code or Domain Id.',
             falcon.HTTP_400],
        13: ['Username Illegal.',
             falcon.HTTP_403],
        14: ['Password Required.',
             falcon.HTTP_400],
        15: ['User ID Required.',
             falcon.HTTP_400],
        16: ['Empty Request. Parameter Need.',
             falcon.HTTP_403],
        17: ['User Duplicated.',
             falcon.HTTP_403],
        18: ["Nothing Happened. The Database was not Changed.",
             falcon.HTTP_403],
        19: ["Username Duplicated.",
             falcon.HTTP_400],
        20: ["Some Parameter is Missing.",
             falcon.HTTP_400],
        21: ["Require Token.",
             falcon.HTTP_400],
        22: ["Login Required.",
             falcon.HTTP_403],
        23: ["Domain Id Required.",
             falcon.HTTP_400],
        24: ["Domain Not Exist.",
             falcon.HTTP_404],
        25: ["Error Username or Password.",
             falcon.HTTP_400],
        26: ["User Not Exist.",
             falcon.HTTP_404],
        27: ["Permission Deny.",
             falcon.HTTP_403],
        28: ["Email Address Illegal.",
             falcon.HTTP_400],
        29: ["Alias Not Exist.",
             falcon.HTTP_404],
        30: ["Error Server Id.",
             falcon.HTTP_400],
        31: ["Request Error. Operate Not Found.",
             falcon.HTTP_400],
        32: ["BCC Not Exist.",
             falcon.HTTP_404],
        33: ["Transport Not Exist.",
             falcon.HTTP_404],
    }


### Invite

Check whether the invite code is available.

    GET /invite?code=123

### Register

Register a new user. You need send username and password in body with invite code in url parameter.

    POST:   /register?code=123
    BODY:   {"username":"catofes","password":"321"}
    
### Login

A hexadecimal token (up to 16 characters, matching the original generator) will be returned if you login successfully. The token will expire in 24h. You need your invite code to change your password.

    #Check if token available.
    GET:    /login?token=1e987d1eba781730
    
    #Login
    POST:   /login
    BODY:   {"username":"catofes","password":"321"}
    RESP:   {"token": "bd36d1ccb2884d6d","username":"catofes","level":1}

    #Logout
    DELETE: /login?token=1e987d1eba781730
    
    #ChangePassword
    PUT:    /login?code=123
    BODY:   {"password":"321"}

### Domains

You can list your domains, add domain or delete domain. When you list a domain or delete it, you need use your domain id instead of number 18.

    #List your domains
    GET:    /domain?token=865d54814424abbe
    RESP:   {"result": [{"id": 16, "name": "sssfsdd.com"}, {"id": 18, "name": "aadd.com"}]}
    
    #Add a domain
    POST:   /domain?token=865d54814424abbe
    BODY:   {"domain": "aadd.com"}
    RESP:   {"result": {"id": 18, "name": "aadd.com"}}
    
    #List a domain 
    GET:    /domain/18?token=865d54814424abbe
    RESP:   {"result": {"id": 18, "name": "aadd.com"}}
    
    #Delete a domain
    DELETE: /domain/18?token=865d54814424abbe
    

### Users

Modify email users belong to a domain. You can list users, add user, delete user or change user's password.
The first number in this api means domain id and the second one means user id.

    #List all user in a domain
    GET:    /user/18?token=865d54814424abbe
    RESP:   {"result": [{"email": "adas@aadd.com", "id": 20, "domain_id": 18}, {"email": "adasss@aadd.com", "id": 21, "domain_id": 18}]}
    
    #Add a user into a domain
    POST:   /user/18?token=865d54814424abbe
    BODY:   {"username": "adas","password": "tatata"}
    
    #List a user
    GET:    /user/18/20?token=865d54814424abbe
    RESP:   {"result": {"email": "adas@aadd.com", "id": 20, "domain_id": 18}}
    
    #Change user's password
    PUT:    /user/18/20?token=865d54814424abbe
    BODY:   {"password": "123"}
    
    #Delete a user
    DELETE: /user/18/20?token=865d54814424abbe
    

### Servers

Show all the mail servers in the system. 

    #List all servers
    GET:    /server?token=b5f1147824e37b8b
    RESP:   {"result": [{"default_mark": "0default", "server_mark": "CNAL", "domain_name": "a.b.com", "region_mark": "3CN"}]
    

### DKIM Settings

You can bind your dkim key to a domain. You can add one or replace one. The first number in this api is domain id. 
When you query your record, not key but sha512(key) will returned.

    #List a dkim record
    GET:    /dkim/1?token=b5f1147824e37b8b
    RESP:   {"key_sha512": "17979e1de7dc2574cc2113a452871e155c78997bd90ae4d03e86ee3d1b210938bc8f0c4f046e0fd715140d026d59c093e6e28a89f2dbed11b1fc3a426e1e832f", "domain": "a.com", "selector": "tau"}

    #Put a dkim record. If you already have one, this api will replace it.
    PUT:    /dkim/1?token=b5f1147824e37b8b
    BODY:   {"selector":"ppp", "private_key":"xxx"}
    
   
### BCC Settings

Show the BCC settings of the mail server. You need level upon 5. Please read postfix manual before add records. 
The region means which servers this record applies to. You can use server mark, region mark or default mark in this field. 
You need put your email username(without @domain.tld) in source field and put a intact email address in destination field.
If you put nothing in your username. All mails sent to your domain will be bcc to your destination.
The first number in the url is domain id and the second one is bcc record id.

    #List all bcc settings
    GET:    /bcc/1?token=b5f1147824e37b8b
    RESP:   {"result": [{"source": "r@aaa.com", "region": "SFDO", "destination": "r+relaycn@aaa.com", "id": 1}]}
    
    #Add a bcc settings
    POST:   /bcc/1?token=b5f1147824e37b8b
    BODY:   {"source":"abc","destination":"bb@gmail.com","region":"SFDO"}
    
    #List a bcc settings
    GET:    /bcc/1/4?token=f0872cbdc28b173
    RESP:   {"result": {"source": "abc@aaa.com", "region": "SFDO", "destination": "bb@gmail.com", "id": 4}}
    
    #Delete a bcc settings
    DELETE: /bcc/1/4?token=f0872cbdc28b173
    
    
### Alias Settings

Shows and modifies alias settings of a domain. Need level 5 above. 
The first number in the url is domain id and the second one is alias record id.

    #List all alias settings
    GET:    /alias/1?token=b5f1147824e37b8b
    RESP:   {"result": [{"source": "sdf@aaa.com", "destination": "haha@gmail.com", "id": 8}]}
    
    #Add a alias settings
    POST:   /alias/1?token=f0872cbdc28b173
    BODY:   {"source":"sdf","destination":"haha@gmail.com"}
    
    #List a alias settings
    GET:    /alias/1/8?token=f0872cbdc28b173
    RESP:   {"result": {"source": "sdf@aaa.com", "destination": "haha@gmail.com", "id": 8}}
    
    #Delete a alias settings
    DELETE: /alias/1/8?token=f0872cbdc28b173

### Transport Settings

Shows and modifies transport settings of a domain. Please read postfix transport manual.
Need level 5 above. You need put "username@" in source field or put nothing in it.

    #List all transport settings
    GET:    /transport/2?token=b5f1147824e37b8b
    RESP:   {"result": [{"source": "aaa.com", "region": "SFDO", "destination": "lmtp:unix:private/dovecot-lmtp", "id": 4}, {"source": "aaa.com", "region": "0default", "destination": "smtp:[xxx.domain.tld]", "id": 5}]}
    
    #Add a transport settings
    POST:   /transport/2?token=f0872cbdc28b173
    BODY:   {"source":"k@","destination":"smtp:[smtp.google.com]","region":"0default"}
    
    #List a transport settings
    GET:    /transport/2/14?token=f0872cbdc28b173
    RESP:   {"result": {"source": "k@aaa.com", "region": "0default", "destination": "smtp:[smtp.google.com]", "id": 14}}
    
    #Delete a transport settings
    DELETE: /transport/2/14?token=f0872cbdc28b173
    
#### Default Transport Settings

Add some default transport settings. This will delete all your transport settings of a domain. Be careful.

    #Get what you can do 
    GET:    /transport_default?token=f0872cbdc28b173
    
    #Do Something. First number is domain id and second number is operate id.
    POST:   /transport_default/1/1?token=f0872cbdc28b173
    
    
    
## 如何使用这套系统

*拥有一个域名。
*获得网站的邀请码。

请联系管理员获取。

*在网站注册一个管理账户。

注意，本管理账户更改密码时候需要您的邀请码。请妥善保管。

*登陆网站并添加一个域名。
*为这个域名添加一个邮箱用户。
*前往您域名的dns服务商并修改您的域名的mx记录。

您可以通过 `server api` 获得系统内邮件服务器的地址。选自一个作为您的mx记录地址即可。

*在您的客户端上设置smtp以及imap/pop3来使用邮箱服务。

服务器地址就是您mx记录添加的地址，用户名为完整的邮箱地址。smtp端口为587协议为STARTSSL，pop3端口为995协议为SSL/TLS，IMAP端口为993协议为SSL/TLS。

### 反垃圾设置。

如果您要是用该系统发送邮件，请设置spf记录以及dkim记录。

在您的dns服务商处添加TXT记录。

    @ TXT v=spf1 include:_spf.catofes.com ~all
    
您可以自己设置您的spf记录。

dkim记录请您自己生成证书后利用api上传，然后添加对应的dns记录即可。

### 转发（别名）设置

请参考postfix中alias的选项。你可以利用api上传您的设置。注意，postfix中别名先于用户处理，即错误的别名设置可能使您的用户收不到任何邮件。

### BCC 设置

类似于别名，BCC应该优先于alias。您可以参考postfix中recipient_bcc选项

### 路由设置

请参考Postfix中transport选项。
