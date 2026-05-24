FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
# Build-time environment variables for Vite
ARG VITE_API=""
ARG VITE_API_KEY=
ENV VITE_API=${VITE_API}
ENV VITE_API_KEY=${VITE_API_KEY}
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
