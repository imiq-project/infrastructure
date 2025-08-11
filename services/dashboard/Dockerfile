FROM node:24
WORKDIR /app/backend
COPY frontend /app/frontend
COPY backend /app/backend
RUN npm install --omit=dev
CMD [ "npm", "run", "server"]
