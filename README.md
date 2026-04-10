# FileArchiver

鏂囦欢褰掓。宸ュ叿锛屾敮鎸?CLI 涓?Web 鐣岄潰锛岃嚜鍔ㄧ増鏈彿鍛藉悕銆佹壒閲忎笂浼犮€佸彲閫?AI 鍒嗙被銆?
## 涓昏鍔熻兘

- CLI 鍗曟枃浠跺綊妗?- Web 鎷栨嫿/澶氶€夋壒閲忎笂浼?- 鑷姩鐗堟湰鍙凤細`-v1`, `-v2`, `-v3`...
- 鍓嶇紑妯℃澘閰嶇疆锛堢増鏈彿鍜屾墿灞曞悕鍥哄畾杩藉姞鍦ㄦ湯灏撅級
- 鍒嗙被鏄犲皠鍙鍖栫紪杈戯紙鏂板/鍒犻櫎锛?- AI 鑷姩鍒嗙被锛堟寜鏂囦欢鍚嶉€愪釜鍒嗙被锛?- 鑷姩鐢熸垚榛樿閰嶇疆鏂囦欢锛坄config.yaml`锛?
## 鐜瑕佹眰

- Go 1.20+

## 蹇€熷紑濮?
### 1. 缂栬瘧

```bash
./scripts/build.sh
```

### 2. 鍚姩 Web

```bash
./scripts/start_web.sh
```

榛樿鍦板潃锛歚http://localhost:8080`

鍙€夛細

```bash
./scripts/start_web.sh :9090 ./config.yaml
```

### 3. CLI 杩愯

```bash
./scripts/run_cli.sh <鏂囦欢璺緞> <鍒嗙被>
```

绀轰緥锛?
```bash
./scripts/run_cli.sh ./report.pdf 寮€鍙?```

## CLI 鍙傛暟

```bash
./archiver -f <鏂囦欢璺緞> -c <鍒嗙被> [-t 鍓嶇紑妯℃澘] [-config 閰嶇疆鏂囦欢]
```

鍙傛暟璇存槑锛?
- `-f`锛氭簮鏂囦欢璺緞锛堝繀濉級
- `-c`锛氬垎绫讳腑鏂囧悕锛堝繀濉紝闇€鍦ㄩ厤缃腑瀛樺湪锛?- `-t`锛氬墠缂€妯℃澘锛堝彲閫夛級
- `-config`锛氶厤缃枃浠惰矾寰勶紙鍙€夛紝榛樿 `./config.yaml`锛?- `-web`锛氬惎鍔?Web
- `-addr`锛歐eb 鐩戝惉鍦板潃锛岄粯璁?`:8080`

## 鍛藉悕瑙勫垯

鏈€缁堟枃浠跺悕鍥哄畾涓猴細

```text
<鍓嶇紑妯℃澘娓叉煋缁撴灉>-v<鐗堟湰鍙?<鍘熸墿灞曞悕>
```

榛樿鍓嶇紑妯℃澘锛?
```text
{category_abbr}-{yyyymmdd}-{filename}
```

鏀寔鍗犱綅绗︼細

- `{category_abbr}`
- `{yyyymmdd}`
- `{filename}`

## Web 浣跨敤璇存槑

- `/`锛氫笂浼犻〉闈紙鏀寔鎷栨嫿/澶氶€夋壒閲忎笂浼狅級
- `/settings`锛氳缃〉闈紙绋嬪簭閰嶇疆 + AI 閰嶇疆锛?
### 鎵归噺涓婁紶鍒嗙被瑙勫垯

- 鍕鹃€?AI锛氭瘡涓枃浠跺垎鍒敱 AI 鍒ゆ柇鍒嗙被
- 涓嶅嬀閫?AI锛氭墍鏈夋枃浠朵娇鐢ㄥ綋鍓嶄笅鎷夋閫変腑鐨勫悓涓€涓垎绫?
## 閰嶇疆鏂囦欢

榛樿閰嶇疆鏂囦欢锛歚./config.yaml`

绋嬪簭鎵句笉鍒伴厤缃枃浠舵椂浼氳嚜鍔ㄥ垱寤洪粯璁ら厤缃€?
### 閰嶇疆绀轰緥

```yaml
archiveBaseDir: archive
templatePrefix: '{category_abbr}-{yyyymmdd}-{filename}'
categories:
  寮€鍙? DEV
  鏁欏: EDU
  璐㈠姟: FIN
ai:
  url: ''
  apiKey: ''
  modelName: ''
```

瀛楁璇存槑锛?
- `archiveBaseDir`锛氬綊妗ｆ牴鐩綍
- `templatePrefix`锛氭枃浠跺悕鍓嶇紑妯℃澘
- `categories`锛氬垎绫绘槧灏勶紙涓枃鍚?-> 缂╁啓锛?- `ai.url` / `ai.apiKey` / `ai.modelName`锛欰I 閰嶇疆锛圤penAI 鍏煎锛?
## 鐩綍缁撴瀯

```text
.
鈹溾攢鈹€ archiver
鈹溾攢鈹€ config.yaml
鈹溾攢鈹€ scripts/
鈹?  鈹溾攢鈹€ build.sh
鈹?  鈹溾攢鈹€ start_web.sh
鈹?  鈹斺攢鈹€ run_cli.sh
鈹溾攢鈹€ internal/
鈹?  鈹溾攢鈹€ archiver/
鈹?  鈹斺攢鈹€ webui/
鈹斺攢鈹€ archive/
```

## 甯歌闂

- 鐩綍閫夋嫨鎸夐挳鏃犳硶杩斿洖瀹屾暣鏈満璺緞锛氭祻瑙堝櫒瀹夊叏闄愬埗瀵艰嚧锛岄〉闈細鏄剧ず鏈嶅姟绔В鏋愬悗鐨勭粷瀵硅矾寰勯瑙堬紱蹇呰鏃跺彲鎵嬪姩濉啓缁濆璺緞銆?- 鍕鹃€?AI 鍚庢姤閿欙細璇峰厛鍦ㄨ缃〉瀹屾暣濉啓 `url`銆乣apiKey`銆乣modelName`銆?
