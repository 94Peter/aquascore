go_package()

go_binary(
    name="aquascore",
    main="main.go",
)

files(
    name="root_files",
    sources=["main.go", "go.mod", "go.sum", "Dockerfile"],
)

docker_image(
    name="image",
    dependencies=[
        ":root_files",
        "//api/cmd:src",
        "//api/internal/crawler:src",
        "//api/internal/crawler/persistence:src",
        "//api/internal/db:src",
        "//api/internal/db/mongo:src",
        "//api/internal/db/mongo/models:src",
        "//api/internal/server:src",
    ],
    image_tags=["latest"],
    repository="94peter/aquascore-api",
)

docker_image(
    name="image-release",
    dependencies=[
        ":root_files",
        "//api/cmd:src",
        "//api/internal/crawler:src",
        "//api/internal/crawler/persistence:src",
        "//api/internal/db:src",
        "//api/internal/db/mongo:src",
        "//api/internal/db/mongo/models:src",
        "//api/internal/server:src",
    ],
    image_tags=["latest"],
    repository="94peter/aquascore-api",
    build_platform=["linux/amd64", "linux/arm64"],
    output={"type": "registry"},
)
