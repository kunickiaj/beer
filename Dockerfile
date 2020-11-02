FROM scratch
LABEL org.opencontainers.image.source https://github.com/kunickiaj/beer

ADD beer /beer

ENTRYPOINT [ "/beer" ]
CMD [ "--help" ]
