window.addEventListener("load", function () {
  window.ui = SwaggerUIBundle({
    url: "./openapi.json",
    dom_id: "#swagger-ui",
    deepLinking: true,
    persistAuthorization: false,
    validatorUrl: null,
    supportedSubmitMethods: ["get", "post", "put", "patch", "delete"],
    presets: [SwaggerUIBundle.presets.apis],
    layout: "BaseLayout",
  });
});
