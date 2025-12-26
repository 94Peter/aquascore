go_package()

go_binary(
    name="bin",
)

files(name="src", sources=["*.go"])



docker_image(
    name="docker",
    dependencies=[
        ":src",
        "api/cmd:src",
        "api/internal/crawler:src",
        "api/internal/crawler/persistence:src",
        "api/internal/db:src",
        "api/internal/db/mongo:src",
        "api/internal/db/mongo/models:src",
        "api/internal/server:src",
        "//:go_files",
    ],
    repository="94peter/aquascore-api",
    image_tags=["latest"],
)

docker_image(
    name="docker-release",
    dependencies=[
        ":src",
        "api/cmd:src",
        "api/internal/crawler:src",
        "api/internal/crawler/persistence:src",
        "api/internal/db:src",
        "api/internal/db/mongo:src",
        "api/internal/db/mongo/models:src",
        "api/internal/server:src",
        "//:go_files",
    ],
    image_tags=["latest"],
    repository="94peter/aquascore-api",
    build_platform=["linux/amd64", "linux/arm64"],
    output={"type": "registry"},
)
