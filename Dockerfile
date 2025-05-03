FROM scratch
COPY relink /
ENTRYPOINT ["/relink"]
