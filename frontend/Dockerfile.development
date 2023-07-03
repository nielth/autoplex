FROM node:18-alpine 

WORKDIR /app

COPY . .

RUN yarn  

ENV NODE_ENV development

EXPOSE 3000

CMD [ "yarn", "start" ]