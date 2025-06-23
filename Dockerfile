FROM python:3.10-slim

# 安装基础依赖（包含编译工具和必要系统库）
RUN apt update && \
    apt install -y build-essential poppler-utils libglib2.0-0 libsm6 libxrender1 libxext6 git && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 安装 mineru（从 Git 安装最新版）
RUN pip install -i https://pypi.tuna.tsinghua.edu.cn/simple git+https://github.com/opendatalab/MinerU.git

# 默认进入 shell，用户可以运行 mineru 命令
CMD ["/bin/bash"]
