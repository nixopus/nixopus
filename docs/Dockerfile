FROM node:20-alpine

WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile

COPY . .
RUN yarn docs:build

EXPOSE 3000

CMD ["yarn", "docs:preview"] 