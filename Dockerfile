FROM scratch
LABEL org.opencontainers.image.source https://github.com/kunickiaj/beer

COPY beer /beer

ENTRYPOINT [ "/beer" ]
CMD [ "--help" ]
