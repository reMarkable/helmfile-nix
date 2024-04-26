{ ... }:
[
  {
    environments = {
      dev = {
        values = [ ];
      };
    };
  }
  {
    releases = [
      {
        name = "test";
        chart = "../chart/";
      }
    ];
  }
]
