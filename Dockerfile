FROM scratch

ADD beer /beer

ENTRYPOINT [ "/beer" ]
CMD [ "--help" ]
