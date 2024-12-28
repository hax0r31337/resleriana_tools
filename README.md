# resleriana_tools
Tool to play around with Atelier Resleriana ( レスレリアーナのアトリエ ) encrypted AssetBundle files

## File Format

### Header
| Name           | Size | Type       | Default | Comment                                                    |  
|----------------|:----:|------------|:-------:|------------------------------------------------------------|
| Magic Number   | 4    | \[\]byte   | "Aktk"  | |
| Version        | 2    | uint16     | 0x01    | |   
| Reserved       | 2    | uint16     | 0x00    | Why does the game checks it must be 0, what's the purpose |
| Encryption     | 4    | enum       | 0x01    | 0x00 => None, 0x01 => Encrypted |
| MD5 Checksum   | 16   | \[\]byte   |         | Hash of the rest of the file whether the file encrypted or not |

### Encryption
A modified version of HChaCha which generates 512-bytes xor block.
Each block generation do 8 normal HChaCha block generation with chained initial state.   
I'll explain details of the format later, you can check out `./encryptor/keygen.go` for details.

## Caution
Unsafe zero-copy type conversion is used to improve performance

## Load from resleriana_tools_config.json and compatible with linux and windows
### First time execute will jump error：
FetchCatalog data: <?xml version="1.0" encoding="UTF-8"?>
<Error><Code>AccessDenied</Code>...</Error>
2024/12/27 20:53:43 Unmarshal error: invalid character '<' looking for beginning of value
2024/12/27 20:53:43 Failed to fetch catalog: invalid character '<' looking for beginning of value

### Bacause resleriana_tools_config.json file will content:
{
  "fetch_url": "https://asset.resleriana.com/asset/AssetsVersion/Android/"
}

### You need replace AssetsVersion to found version string by yourself.
Then execute again without compile.
