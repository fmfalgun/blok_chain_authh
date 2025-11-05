# IoT Device Simulator Dockerfile

FROM node:16-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install --production

# Copy application files
COPY . .

# Create wallet directory
RUN mkdir -p wallet

# Set environment variables
ENV NODE_ENV=production

# Run as non-root user
RUN addgroup -g 1001 -S iotuser && \
    adduser -u 1001 -S iotuser -G iotuser && \
    chown -R iotuser:iotuser /app

USER iotuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD node -e "console.log('healthy')" || exit 1

# Start simulator
CMD ["npm", "start"]
