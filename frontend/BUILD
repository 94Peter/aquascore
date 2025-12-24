# Define a target for the Docker image of the frontend application
# Optimized for local development on Apple Silicon (arm64)
docker_image(
    name="image",
    repository="94peter/aquascore-frontend",
    image_tags=["latest"],
    dependencies=[
        ":frontend-docker-context",
    ],
    # Force native architecture to avoid QEMU emulation issues (exit code 254)
    build_platform=["linux/arm64"], 
)

# Define a release target for multi-platform support
docker_image(
    name="image-release",
    repository="94peter/aquascore-frontend",
    image_tags=["latest"],
    dependencies=[
        ":frontend-docker-context",
    ],
    build_platform=["linux/amd64", "linux/arm64"],
    output={"type": "registry"},
)

# Define a single 'files' target to ensure all necessary files are included in the Docker build context
# and available at the root of the context for the Dockerfile.
files(
    name="frontend-docker-context",
    sources=[
        "package.json", 
        "package-lock.json", # Re-add package-lock.json as it should be generated correctly now
        "vite.config.ts", 
        "tsconfig.json", 
        "tsconfig.app.json", 
        "tsconfig.node.json", 
        "tailwind.config.ts", 
        "postcss.config.js", 
        "eslint.config.js", 
        "index.html", 
        "README.md", 
        ".gitignore",
        "public/**", 
        "src/**", # Include all source files
        "Dockerfile", # Include the Dockerfile itself in the context
        "nginx.conf.template", # Include the Nginx template
    ],
)