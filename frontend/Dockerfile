# Stage 1: Build the React application
FROM --platform=linux/arm64 node:20-slim as builder

WORKDIR /app

# Copy package.json and package-lock.json to install dependencies
COPY frontend/package.json ./
COPY frontend/package-lock.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application code
COPY frontend/. .

# Build the React app
# The 'build' script is defined in package.json and typically runs 'vite build'
RUN npm run build


# Stage 2: Serve the application with Nginx
FROM --platform=linux/arm64 nginx:alpine

# Install envsubst which is needed for runtime variable substitution
RUN apk add --no-cache gettext

# Copy the built React app from the builder stage to Nginx's public directory
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy the Nginx configuration template
COPY frontend/nginx.conf.template /etc/nginx/templates/default.conf.template

# Use envsubst to replace the placeholder at runtime
CMD ["/bin/sh", "-c", "envsubst '${API_TARGET}' < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]


