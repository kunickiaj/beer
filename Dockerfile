FROM scratch
LABEL org.opencontainers.image.source https://github.com/kunickiaj/beer

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/beer /beer

ENTRYPOINT [ "/beer" ]
CMD [ "--help" ]
