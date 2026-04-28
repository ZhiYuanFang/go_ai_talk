#指定基础镜像
FROM loads/alpine:3.8 

###############################################################################
#                                INSTALLATION
###############################################################################
#在镜像中声明环境变量，理解为一个常量定义
ENV WORKDIR                 /app
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
#把resource文件添加到基础镜像，文件保存路径指定在$WORKDIR中
ADD resource                $WORKDIR/resource
#把./temp/linux_amd64/main文件添加到基础镜像，文件保存路径指定在$WORKDIR/main中
ADD ./main $WORKDIR/main
#在基础镜像的环境中执行命令 使$WORKDIR/main 文件开放可读、写、执行，并生成一个镜像层
RUN chmod +x $WORKDIR/main

###############################################################################
#                                   START
###############################################################################
#以下的指令的工作目录为：$WORKDIR
WORKDIR $WORKDIR
#镜像中的默认执行文件：./main
CMD ./main
