FROM alpine:3.17 AS builder

RUN apk add --no-cache curl
RUN curl -O https://pkgs.tailscale.com/stable/tailscale_1.90.8_arm.tgz
RUN tar -xvf tailscale_1.90.8_arm.tgz

FROM balenalib/raspberry-pi-alpine:3.17-run

COPY --from=builder /tailscale_1.90.8_arm/tailscaled /usr/bin/tailscaled
COPY --from=builder /tailscale_1.90.8_arm/tailscale /usr/bin/tailscale
COPY tailscale/init.sh /init.sh

CMD /init.sh
