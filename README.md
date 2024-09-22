# Lifting Gopher

Lifting Gopher is an interactive web application that uses WebAssembly and Go to provide a fun and engaging experience. The application captures video from your webcam and overlays a lifting gopher animation.

## Demo
Check out the live demo: [Lifting Gopher Demo Page](https://lifting-gopher.pages.dev)

https://github.com/user-attachments/assets/db427a0e-2bdc-47df-a8b7-ca2a258fe40f

## Features

- Real-time video capture
- Fun gopher animation overlay
- Built with Go and WebAssembly(using Ebitengine)

## Installation

To run the project locally, follow these steps:

1. **Clone the repository:**

    ```sh
    git clone https://github.com/ponyo877/lifting-gopher.git
    cd lifting-gopher
    ```

2. **Install dependencies:**

    Ensure you have Go installed. Then, install the necessary Go packages:

    ```sh
    go mod tidy
    ```

3. **Build the WebAssembly binary:**

    ```sh
    GOOS=js GOARCH=wasm go build -o main.wasm
    ```

4. **Serve the application:**

    You can use any static file server to serve the `index.html` file. For example, using `http-server`:

    ```sh
    # if you use npx
    npx http-server .
    ```

5. **Open the application:**

    Open your browser and navigate to `http://localhost:8080` (or the port your server is running on).

## Usage

Once the application is running, allow access to your webcam. You should see the video feed with the gopher animation overlay. Use the buttons to interact with the gopher.


## Acknowledgements

- [Ebitengine](https://github.com/hajimehoshi/ebiten) - A dead simple 2D game library for Go

---

Thank you for checking out Lifting Gopher! We hope you enjoy using it as much as we enjoyed building it.
