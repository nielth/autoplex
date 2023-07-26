FROM node:18-alpine 

WORKDIR /app

COPY . .

RUN yarn  

RUN yarn build

RUN yarn global add serve

ENV NODE_ENV production

EXPOSE 3000

CMD [ "serve", "-s", "build" ]