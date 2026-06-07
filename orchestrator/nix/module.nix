{ config, lib, pkgs, ... }:

let
  cfg = config.services.octra-orchestrator;
in {
  options.services.octra-orchestrator = {
    enable = lib.mkEnableOption "Octra orchestrator service";

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.octra-orchestrator;
      description = "Orchestrator package to use";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 50052;
      description = "gRPC server port";
    };

    environment = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      default = {
        ORCHESTRATOR_GRPC_PORT = "50052";
        PROJECTS_DIR = "/var/lib/octra/projects";
        DB_DNS = "postgres://octra:octra@localhost:5432/octra?sslmode=disable";
        REDIS_URL = "redis://localhost:6379/0";
        AGENTS_SERVICE_HOST = "localhost:50053";
      };
      description = "Environment variables for the orchestrator";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.octra-orchestrator = {
      description = "Octra AI orchestrator";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" "postgresql.service" "redis.service" ];

      path = [ pkgs.nix ];

      serviceConfig = {
        ExecStart = "${cfg.package}/bin/app";
        Restart = "on-failure";
        RestartSec = "5s";
        User = "octra";
        Group = "octra";
        StateDirectory = "octra";
        WorkingDirectory = "/var/lib/octra";
        Environment = cfg.environment;
      };
    };

    users.users.octra = {
      isSystemUser = true;
      group = "octra";
      home = "/var/lib/octra";
      createHome = true;
    };

    users.groups.octra = {};

    nix.settings.trusted-users = [ "octra" ];
  };
}
