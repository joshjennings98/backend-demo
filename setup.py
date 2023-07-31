from setuptools import setup, find_packages

setup(
    name='backend_demo',
    version='0.1.0',
    packages=['backend_demo'],
    package_data={
        'backend_demo': ['templates/*', 'static/*'],
    },
    entry_points={
        'console_scripts': [
            'backend_demo = backend_demo.__main__:main',
        ],
    },
    install_requires=[
        'flask',
    ],
)
