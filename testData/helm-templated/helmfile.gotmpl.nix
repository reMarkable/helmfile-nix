{ escape_var, ... }:
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
        hooks = [
          {
            events = [
              "presync"
              "prepare"
            ];
            showlogs = true;
            command = "echo";
            args = [
              "--environment"
              (escape_var "{{ .Environment | toJson }}")
              "--release"
              (escape_var "{{ .Release | toJson }}")
              "--event"
              (escape_var "{{ .Event | toJson }}")
            ];
          }
        ];
      }
    ];
  }
]
