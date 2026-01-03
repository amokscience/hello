# Use an official Go image to build
FROM golang:1.25-alpine AS build

WORKDIR /app

# Copy source
COPY . .

# Build a static binary
RUN go build -o hello .

# Use a small final image
FROM alpine:3.20

WORKDIR /app

# Copy only the binary from the build stage
COPY --from=build /app/hello .
COPY --from=build /app/input.txt ./
COPY --from=build /app/settings.json ./

# Expose port
EXPOSE 8080

# Run the application
CMD ["./hello"]

# Command to run the program
ENTRYPOINT ["./hello"]
