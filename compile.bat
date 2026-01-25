@echo off

go build -o luffy-windows-amd64.exe
upx --best --lzma luffy-windows-amd64.exe
