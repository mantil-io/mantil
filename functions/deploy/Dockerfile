FROM public.ecr.aws/lambda/go:latest

# terrraform install
RUN yum -y install unzip wget git
ARG version=1.0.0
ARG arch=_linux_amd64 #_linux_arm64
RUN wget https://releases.hashicorp.com/terraform/$version/terraform_${version}${arch}.zip \
    && unzip terraform_${version}${arch}.zip \
    && rm terraform_${version}${arch}.zip \
    && mv terraform /usr/bin/ \
    && terraform --version
ENV TF_IN_AUTOMATION=1

# aws cli install
# ref: https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html#cliv2-linux-install
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" \
    && unzip awscliv2.zip \
    && ./aws/install

COPY bootstrap /var/task/
CMD [ "bootstrap" ]
