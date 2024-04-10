FROM scratch
COPY ctfd-setup /
ENTRYPOINT [ "/ctfd-setup" ]
