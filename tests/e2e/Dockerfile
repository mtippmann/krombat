FROM mcr.microsoft.com/playwright:v1.58.2-noble
WORKDIR /work
COPY package.json package-lock.json ./
RUN npm ci
COPY tests/e2e ./tests/e2e
CMD ["node", "tests/e2e/smoke-test.js"]
