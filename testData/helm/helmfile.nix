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
      {
        name = "testNix";
        nixChart = "../chart/";
      }
    ];
  }
]
