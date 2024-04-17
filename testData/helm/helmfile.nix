{ ... }: [
  { environments = { dev = { values = [ ]; }; }; }
  {
    repositories = [{
      name = "grafana";
      url = "https://grafana.github.io/helm-charts";
    }];
    releases = [{
      name = "grafana";
      chart = "grafana/grafana";
    }];
  }
]
