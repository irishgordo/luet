FROM golang as builder
ADD . /luet
RUN cd /luet && make build

FROM scratch
ENV LUET_NOLOCK=true
COPY --from=builder /luet/luet /usr/bin/luet

ENTRYPOINT ["/usr/bin/luet"]
