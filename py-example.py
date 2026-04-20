      @validator.check
      def load_palette(self, path: str) -> Config:
          """Load and validate a TOML palette."""
          with open(path, "rb") as f:
              self.data = tomllib.load(f)
          if not self.validate():
              raise PaletteError(f"Invalid: {path}")
          return self.config

