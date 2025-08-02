> **[↩ 返回到目录](doc.md)**

# Go 标准库参考文档

## archive
### tar
tar 包实现了对 tar 归档文件的访问。

### zip
zip 包提供了对读写 ZIP 归档文件的支持。

## bufio
bufio 包实现了带缓冲的 I/O。它包装一个 `io.Reader` 或 `io.Writer` 对象，创建另一个也实现了该接口的对象（`Reader` 或 `Writer`），但提供了缓冲和一些文本 I/O 的辅助功能。

## builtin
builtin 包为 Go 的预声明标识符提供文档。

## bytes
bytes 包实现了用于操作字节切片的函数。

## cmp
cmp 包提供了与比较有序值相关的类型和函数。

## compress
### bzip2
bzip2 包实现了 bzip2 解压缩。

### flate
flate 包实现了 DEFLATE 压缩数据格式，如 RFC 1951 中所述。

### gzip
gzip 包实现了读写 gzip 格式的压缩文件，如 RFC 1952 中所规定。

### lzw
lzw 包实现了 Lempel-Ziv-Welch 压缩数据格式，如 T. A. Welch 的《高性能数据压缩技术》中所述。

### zlib
zlib 包实现了读写 zlib 格式的压缩数据，如 RFC 1950 中所规定。

## container
### heap
heap 包为任何实现 `heap.Interface` 的类型提供堆操作。

### list
list 包实现了双向链表。

### ring
ring 包实现了对循环链表的操作。

## context
context 包定义了 `Context` 类型，该类型在 API 边界和进程之间传递截止时间、取消信号和其他请求范围的值。

## crypto
crypto 包收集了常见的加密常量。

### aes
aes 包实现了美国联邦信息处理标准出版物 197 中定义的 AES 加密（前称 Rijndael）。

### cipher
cipher 包实现了标准的分组密码模式，可以包装在低级的分组密码实现之上。

### des
des 包实现了美国联邦信息处理标准出版物 46-3 中定义的数据加密标准 (DES) 和三重数据加密算法 (TDEA)。

### dsa
dsa 包实现了 FIPS 186-3 中定义的数字签名算法。

### ecdh
ecdh 包实现了 NIST 曲线和 Curve25519 上的椭圆曲线迪菲-赫尔曼密钥交换。

### ecdsa
ecdh 包实现了 [FIPS 186-5] 中定义的椭圆曲线数字签名算法。

### ed25519
ed25519 包实现了 Ed25519 签名算法。

### elliptic
elliptic 包实现了标准的 NIST P-224, P-256, P-384 和 P-521 素域椭圆曲线。

### fips140

### hkdf
hkdf 包实现了 RFC 5869 中定义的基于 HMAC 的提取和扩展密钥派生函数 (HKDF)。

### hmac
hmac 包实现了美国联邦信息处理标准出版物 198 中定义的密钥散列消息认证码 (HMAC)。

### md5
md5 包实现了 RFC 1321 中定义的 MD5 哈希算法。

### mlkem
mlkem 包实现了量子抗性密钥封装方法 ML-KEM（前称 Kyber），如 [NIST FIPS 203] 中所规定。

### pbkdf2
pbkdf2 包实现了 RFC 8018 (PKCS #5 v2.1) 中定义的密钥派生函数 PBKDF2。

### rand
rand 包实现了密码学安全的随机数生成器。

### rc4
rc4 包实现了 Bruce Schneier 的《应用密码学》中定义的 RC4 加密。

### rsa
rsa 包实现了 PKCS #1 和 RFC 8017 中规定的 RSA 加密。

### sha1
sha1 包实现了 RFC 3174 中定义的 SHA-1 哈希算法。

### sha256
sha256 包实现了 FIPS 180-4 中定义的 SHA224 和 SHA256 哈希算法。

### sha3
sha3 包实现了 FIPS 202 中定义的 SHA-3 哈希算法和 SHAKE 可扩展输出函数。

### sha512
sha512 包实现了 FIPS 180-4 中定义的 SHA-384, SHA-512, SHA-512/224, 和 SHA-512/256 哈希算法。

### subtle
subtle 包实现了在加密代码中经常有用但需要谨慎使用才能正确使用的函数。

### tls
tls 包部分实现了 RFC 5246 中规定的 TLS 1.2 和 RFC 8446 中规定的 TLS 1.3。

### x509
x509 包实现了 X.509 标准的一个子集。

### x509/pkix
pkix 包包含用于 X.509 证书、CRL 和 OCSP 的 ASN.1 解析和序列化的共享低级结构。

## database
### sql
sql 包为 SQL（或类 SQL）数据库提供了通用接口。

### sql/driver
driver 包定义了数据库驱动程序需要实现的接口，供 sql 包使用。

## debug
### buildinfo
buildinfo 包提供了访问嵌入在 Go 二进制文件中关于其构建方式信息的功能。

### dwarf
dwarf 包提供了访问从可执行文件加载的 DWARF 调试信息的功能，如 DWARF 2.0 标准中所定义。

### elf
elf 包实现了对 ELF 目标文件的访问。

### gosym
gosym 包实现了对 gc 编译器生成的 Go 二进制文件中嵌入的 Go 符号和行号表的访问。

### macho
macho 包实现了对 Mach-O 目标文件的访问。

### pe
pe 包实现了对 PE（微软 Windows 可移植可执行）文件的访问。

### plan9obj
plan9obj 包实现了对 Plan 9 a.out 目标文件的访问。

## embed
embed 包提供了访问嵌入在运行中的 Go 程序中的文件的功能。

## encoding
encoding 包定义了被其他包共享的接口，这些包用于在字节级和文本表示之间转换数据。

### ascii85
ascii85 包实现了 btoa 工具以及 Adobe 的 PostScript 和 PDF 文档格式中使用的 ascii85 数据编码。

### asn1
asn1 包实现了 ITU-T Rec X.690 中定义的 DER 编码 ASN.1 数据结构的解析。

### base32
base32 包实现了 RFC 4648 中规定的 base32 编码。

### base64
base64 包实现了 RFC 4648 中规定的 base64 编码。

### binary
binary 包实现了数字和字节序列之间的简单转换，以及变长整数的编码和解码。

### csv
csv 包读写逗号分隔值 (CSV) 文件。

### gob
gob 包管理 gobs 流——在编码器（发送方）和解码器（接收方）之间交换的二进制值。

### hex
hex 包实现了十六进制编码和解码。

### json
json 包实现了 RFC 7159 中定义的 JSON 的编码和解码。

### pem
pem 包实现了 PEM 数据编码，起源于"隐私增强邮件"。

### xml
xml 包实现了一个简单的 XML 1.0 解析器，能够理解 XML 名称空间。

## errors
errors 包实现了用于操作错误的函数。

## expvar
expvar 包为公共变量（如服务器中的操作计数器）提供了标准化接口。

## flag
flag 包实现了命令行标志解析。

## fmt
fmt 包实现了格式化的 I/O，其函数类似于 C 语言的 printf 和 scanf。

## go
### ast
ast 包声明了用于表示 Go 包语法树的类型。

### build
build 包收集有关 Go 包的信息。

### build/constraint
build/constraint 包实现了构建约束行的解析和求值。

### constant
constant 包实现了表示无类型 Go 常量的值及其相应操作。

### doc
doc 包从 Go 抽象语法树 (AST) 中提取源代码文档。

### doc/comment
doc/comment 包实现了 Go 文档注释（即紧邻包、const、func、type 或 var 顶层声明的注释）的解析和重新格式化。

### format
format 包实现了 Go 源代码的标准格式化。

### importer
importer 包提供了对导出数据导入器的访问。

### parser
parser 包实现了 Go 源文件的解析器。

### printer
printer 包实现了 AST 节点的打印。

### scanner
scanner 包实现了 Go 源代码文本的扫描器。

### token
token 包定义了表示 Go 编程语言词法标记的常量以及对标记的基本操作（打印、谓词）。

### types
types 包声明了数据类型并实现了 Go 包的类型检查算法。

### version
version 包提供了对 [Go 版本] 在 [Go 工具链名称语法] 中的操作：如 "go1.20", "go1.21.0", "go1.22rc2", 和 "go1.23.4-bigcorp" 这样的字符串。

## hash
hash 包提供了哈希函数的接口。

### adler32
adler32 包实现了 Adler-32 校验和。

### crc32
crc32 包实现了 32 位循环冗余校验（CRC-32）校验和。

### crc64
crc64 包实现了 64 位循环冗余校验（CRC-64）校验和。

### fnv
fnv 包实现了由 Glenn Fowler, Landon Curt Noll, 和 Phong Vo 创建的非加密哈希函数 FNV-1 和 FNV-1a。

### maphash
maphash 包提供了对字节序列和可比较值的哈希函数。

## html
html 包提供了对 HTML 文本进行转义和反转义的函数。

### template
template (html/template) 包实现了数据驱动的模板，用于生成防代码注入的 HTML 输出。

## image
image 包实现了一个基本的二维图像库。

### color
color 包实现了一个基本的颜色库。

### color/palette
palette 包提供了标准颜色调色板。

### draw
draw 包提供了图像合成函数。

### gif
gif 包实现了 GIF 图像解码器和编码器。

### jpeg
jpeg 包实现了 JPEG 图像解码器和编码器。

### png
png 包实现了 PNG 图像解码器和编码器。

## index
### suffixarray
suffixarray 包实现了在内存后缀数组上以对数时间进行子字符串搜索。

## io
io 包提供了对 I/O 原语的基本接口。

### fs
fs 包定义了对文件系统的基本接口。

### ioutil
ioutil 包实现了一些 I/O 实用函数。

### iter
iter 包提供了与序列迭代器相关的基本定义和操作。

## log
log 包实现了一个简单的日志记录包。

### slog
slog 包提供了结构化日志记录，其中日志记录包含一条消息、一个严重性级别和各种以键值对形式表示的其他属性。

### syslog
syslog 包为系统日志服务提供了一个简单接口。

## maps
maps 包定义了各种对任何类型的映射 (map) 都有用的函数。

## math
math 包提供了基本的数学常量和函数。

### big
big 包实现了任意精度的算术（大数）。

### bits
bits 包为预声明的无符号整数类型实现了位计数和操作函数。

### cmplx
cmplx 包为复数提供了基本的数学常量和函数。

### rand
rand 包实现了适用于模拟等任务的伪随机数生成器，但不应将其用于安全敏感的工作。

### rand/v2
rand 包实现了适用于模拟等任务的伪随机数生成器，但不应将其用于安全敏感的工作。

## mime
mime 包实现了 MIME 规范的部分内容。

### multipart
multipart 包实现了如 RFC 2046 中定义的 MIME 多部分解析。

### quotedprintable
quotedprintable 包实现了 RFC 2045 中规定的引号可打印编码。

## net
net 包为网络 I/O 提供了可移植的接口，包括 TCP/IP、UDP、域名解析和 Unix 域套接字。

### http
http 包提供了 HTTP 客户端和服务器的实现。

### http/cgi
cgi 包实现了 RFC 3875 中规定的 CGI（通用网关接口）。

### http/cookiejar
cookiejar 包实现了符合 RFC 6265 的内存中的 `http.CookieJar`。

### http/fcgi
fcgi 包实现了 FastCGI 协议。

### http/httptest
httptest 包为 HTTP 测试提供了工具。

### http/httptrace
httptrace 包提供了追踪 HTTP 客户端请求内事件的机制。

### http/httputil
httputil 包提供了 HTTP 实用函数，补充了 `net/http` 包中更常用的函数。

### http/pprof
pprof 包通过其 HTTP 服务器以 pprof 可视化工具期望的格式提供运行时性能分析数据。

### mail
mail 包实现了邮件消息的解析。

### netip
netip 包定义了一个 IP 地址类型，它是一个小型的值类型。

### rpc
rpc 包提供了通过网络或其他 I/O 连接访问对象导出方法的功能。

### rpc/jsonrpc
jsonrpc 包为 rpc 包实现了 JSON-RPC 1.0 ClientCodec 和 ServerCodec。

### smtp
smtp 包实现了 RFC 5321 中定义的简单邮件传输协议。

### textproto
textproto 包实现了 HTTP、NNTP 和 SMTP 风格的基于文本的请求/响应协议的通用支持。

### url
url 包解析 URL 并实现查询转义。

## os
os 包提供了对操作系统功能的平台无关接口。

### exec
exec 包运行外部命令。

### signal
signal 包实现了对传入信号的访问。

### user
user 包允许通过名称或 ID 查找用户账户。

## path
### path
path 包实现了用于操作斜杠分隔路径的实用程序例程。

### filepath
filepath 包实现了以与目标操作系统定义的文件路径兼容的方式操作文件名路径的实用程序例程。

## plugin
plugin 包实现了 Go 插件的加载和符号解析。

## reflect
reflect 包实现了运行时反射，允许程序操作具有任意类型的对象。

## regexp
regexp 包实现了正则表达式搜索。

### syntax
syntax 包将正则表达式解析为解析树，并将解析树编译为程序。

## runtime
runtime 包包含与 Go 运行时系统交互的操作，例如控制 goroutine 的函数。

### cgo
cgo 包包含 cgo 工具生成的代码的运行时支持。

### coverage
coverage 包包含在运行时从长时间运行和/或服务器程序中写入覆盖率分析数据的 API，这些程序不通过 `os.Exit` 终止。

### debug
debug 包包含了程序在运行时调试自身的功能。

### metrics
metrics 包提供了访问 Go 运行时导出的实现定义指标的稳定接口。

### pprof
pprof 包以 pprof 可视化工具期望的格式写入运行时性能分析数据。

### race
race 包实现了数据竞争检测逻辑。

### trace
trace 包包含了程序生成 Go 执行追踪器追踪的功能。

## slices
slices 包定义了各种对任何类型的切片都有用的函数。

## sort
sort 包为排序切片和用户定义的集合提供了基础原语。

## strconv
strconv 包实现了基本数据类型的字符串表示形式的转换。

## strings
strings 包实现了用于操作 UTF-8 编码字符串的简单函数。

## structs
structs 包定义了可以作为结构体字段使用的标记类型，以修改结构体的属性。

## sync
sync 包提供了基本的同步原语，例如互斥锁。

### atomic
atomic 包提供了低级的原子内存原语，对实现同步算法很有用。

## syscall
syscall 包包含对低级操作系统原语的接口。

## js
### js
js 包在使用 js/wasm 架构时，提供了对 WebAssembly 主机环境的访问。

## testing
testing 包为 Go 包的自动化测试提供了支持。

### fstest
fstest 包为测试文件系统实现和使用者提供了支持。

### iotest
iotest 包实现了主要用于测试的 Reader 和 Writer。

### quick
quick 包实现了帮助进行黑盒测试的实用函数。

### slogtest
slogtest 包为测试 `log/slog.Handler` 的实现提供了支持。

### synctest
synctest 包为测试并发代码提供了支持。

## text
### scanner
scanner 包为 UTF-8 编码的文本提供了扫描器和分词器。

### tabwriter
tabwriter 包实现了一个写入过滤器（`tabwriter.Writer`），将输入中的制表符分隔的列转换为正确对齐的文本。

### template
template 包实现了数据驱动的模板，用于生成文本输出。

### template/parse
parse 包为 `text/template` 和 `html/template` 定义的模板构建解析树。

## time
time 包提供了测量和显示时间的功能。

### tzdata
tzdata 包提供了一个嵌入式的时区数据库副本。

## unicode
unicode 包提供了数据和函数，用于测试 Unicode 码点的某些属性。

### utf16
utf16 包实现了 UTF-16 序列的编码和解码。

### utf8
utf8 包实现了支持 UTF-8 编码文本的函数和常量。

## unique
unique 包提供了对可比较值进行规范化（"驻留"）的工具。

## unsafe
unsafe 包包含了绕过 Go 程序类型安全的操作。

## weak
weak 包提供了安全地弱引用内存（即，不会阻止其回收）的方法。

> **[↩ 返回到目录](doc.md)**
