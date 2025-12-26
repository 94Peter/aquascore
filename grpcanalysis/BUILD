python_sources(name="lib")


pex_binary(
    name="pex_binary",
    entry_point="server.py",
    dependencies=[
        ":lib",
        "grpcanalysis/logic:lib",
        "//:reqs",
    ],
)


files(
    name="root_files",
    sources=[
        "server.py",
        "__init__.py",
    ],
)

docker_image(
    name="docker",
    dependencies=[
        ":root_files",
        "grpcanalysis/logic:src",
        "//:pythons_files"
    ],
    image_tags=["latest"],
    repository="94peter/aquascore-analysis",
)

docker_image(
    name="docker-release",
    dependencies=[
        ":root_files",
        "grpcanalysis/logic:src",
        "//:pythons_files"
    ],
    image_tags=["latest"],
    repository="94peter/aquascore-analysis",
    build_platform=["linux/amd64", "linux/arm64"],
    output={"type": "registry"}, 
)
