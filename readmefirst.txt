
./geektime-downloader --gcess XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX --gcid XXXXXXXXXXXXXXXXXX -f /media/XXXXX/XXXXX/geektime/java --quiet --product-type 0 --interval 3 --product-id 100047701,100023901


(this is a mark for me only) if want to uninstall go, just remove these pkgs:
golang-1.22-go golang-1.22-src golang-go golang-src


TODO:
在大厂案例下载代码中，额外请求 /serv/v3/article/info，取得其中的 data.info.attachments.url, 下载文件并保存为 data.info.attachments.filename
另外，下载一下文稿（文本稿），路径是 data.info.video.manuscripts (数组)