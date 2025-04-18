FROM node:20-alpine

WORKDIR /app

RUN apk add --no-cache git && \
    git config --global user.email "build@nixopus.com" && \
    git config --global user.name "Nixopus Build"

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile

COPY . .
RUN git init && \
    git add . && \
    git commit -m "Initial commit" --allow-empty

RUN yarn docs:build

EXPOSE 3000

CMD ["yarn", "docs:preview"] 