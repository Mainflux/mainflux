import setuptools

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setuptools.setup(
    name="Mainflux",
    version="0.0.1",
    author="Filip Bugarski",
    author_email="filipbugarski@gmail.com",
    description="Python SDK",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/mainflux/mainflux/tree/master/pkg/sdk/python",
    project_urls={
        "Bug Tracker": "https://github.com/mainflux/mainflux/tree/master/pkg/sdk/python",
    },
    classifiers=[
        "Programming Language :: Python :: 3.8.5",
        "Operating System :: OS Independent",
    ],
    package_dir={"": "src"},
    packages=setuptools.find_packages(where="src"),
    python_requires=">=3.6",
)
