ARG COMMIT_SHA
# using base image created by kind https://github.com/kubernetes-sigs/kind/blob/master/images/base/Dockerfile
# which is an ubuntu 19.10 with an entry-point that helps running systemd
# could be changed to any debian that can run systemd
FROM gcr.io/k8s-minikube/kindbase:v0.0.10-snapshot as base
USER root
# specify version of everything explicitly using 'apt-cache policy'
RUN clean-install \
    lz4 \
    gnupg \
    sudo \
    openssh-server \
    dnsutils

RUN clean-install containerd

RUN clean-install docker.io

# install cri-o based on https://github.com/cri-o/cri-o/commit/96b0c34b31a9fc181e46d7d8e34fb8ee6c4dc4e1#diff-04c6e90faac2675aa89e2176d2eec7d8R128
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_19.10/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list" && \    
    curl -LO https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable/xUbuntu_19.10/Release.key && \
    apt-key add - < Release.key && apt-get update && \
    clean-install cri-o-1.17 cri-tools

# install podman
# install varlink
# libglib2.0-0 is required for conmon, which is required for podman
RUN clean-install podman varlink libglib2.0-0

# disable non-docker runtimes by default
RUN systemctl disable containerd && systemctl disable crio && rm /etc/crictl.yaml
# enable docker which is default
RUN systemctl enable docker
# making SSH work for docker container 
# based on https://github.com/rastasheep/ubuntu-sshd/blob/master/18.04/Dockerfile
RUN mkdir /var/run/sshd
RUN echo 'root:root' |chpasswd
RUN sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config
EXPOSE 22
# create docker user for minikube ssh. to match VM using "docker" as username
RUN adduser --ingroup docker --disabled-password --gecos '' docker 
RUN adduser docker sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER docker
RUN mkdir /home/docker/.ssh
USER root
# kind base-image entry-point expects a "kind" folder for product_name,product_uuid
# https://github.com/kubernetes-sigs/kind/blob/master/images/base/files/usr/local/bin/entrypoint
RUN mkdir -p /kind
RUN echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"
