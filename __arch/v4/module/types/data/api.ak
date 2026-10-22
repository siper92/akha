//@API:DEF
ak.import(rest)

rest.setup(
    "localhost",
    "8080",
);

rest.get("/api/v1/data", {
  data: {
    message: "Hello, World!"
  }
})