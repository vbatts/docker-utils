FROM busybox
COPY . /
# TODO get gpg in the image
ENTRYPOINT sha256sum -c /dockertarsum.sum && /dockertarsum
