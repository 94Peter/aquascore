python_sources(name="lib")

python_requirements(name="reqs")

files(
    name="root_files",
    sources=[
        "requirements.txt",
        "server.py",
        "__init__.py",
    ],
)

docker_image(
    name="image",
    dependencies=[
        ":root_files",
        "analysis/logic:src",
    ],
    image_tags=["latest"],
    repository="94peter/aquascore-analysis",
)

docker_image(
    name="image-release",
    dependencies=[
        ":root_files",
        "analysis/logic:src",
    ],
    image_tags=["latest"],
    repository="94peter/aquascore-analysis",
    build_platform=["linux/amd64", "linux/arm64"],
    output={"type": "registry"}, 
)
