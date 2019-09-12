FROM cockroachdb/cockroach:v19.1.4

ENV COCKROACH_USER=schemahero
ENV COCKROACH_DATABASE=schemahero

COPY cockroach.sh /cockroach/

ENTRYPOINT ["/cockroach/cockroach.sh"]
CMD ["start", "--insecure"]
