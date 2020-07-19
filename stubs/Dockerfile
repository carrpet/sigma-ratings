FROM node:10-alpine

RUN npm install --global json-server
EXPOSE 3000
WORKDIR /stubs
ENTRYPOINT ["json-server"]
CMD ["--host", "0.0.0.0", "-p", "80", "-w", "db.json"]
