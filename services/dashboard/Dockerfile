FROM node:24 AS builder
WORKDIR /app/frontend
COPY frontend /app/frontend
RUN npm install && npm run build

FROM node:24
WORKDIR /app/backend
COPY --from=builder /app/frontend/dist /app/frontend/dist
COPY backend /app/backend
RUN npm install --omit=dev
CMD [ "npm", "run", "server"]
