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