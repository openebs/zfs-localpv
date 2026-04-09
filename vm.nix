{ ... }:
let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
in
{
  nix.nixPath = [
    "nixpkgs=${pkgs.path}"
  ];
  nixos-shell.mounts = {
    mountHome = false;
    mountNixProfile = false;
    cache = "none"; # default is "loose"

    extraMounts = {
      "/zfs" = {
        target = ./.;
        cache = "none";
      };
      ${if (builtins.getEnv "TERMINFO" != "") then "/terminfo" else null } = {
        target = builtins.getEnv "TERMINFO";
      };
    };
  };

  virtualisation = {
    cores = 4;
    memorySize = 2048;
    # Uncomment to be able to ssh into the vm, example:
    # ssh -p 2222 -o StrictHostKeychecking=no root@localhost
    # forwardPorts = [
    #  { from = "host"; host.port = 2222; guest.port = 22; }
    # ];
    diskSize = 20 * 1024;
    docker = {
      enable = true;
    };
  };
  documentation.enable = false;

  boot = {
    supportedFilesystems = [ "zfs" ];
    zfs.forceImportRoot = false;
    zfs.package = pkgs.zfs;
  };

  networking = {
    firewall = {
      allowedTCPPorts = [
        6443 # k3s: required so that pods can reach the API server (running on port 6443 by default)
      ];
    };

    # Set from machine-id or `head -c4 /dev/urandom | od -A none -t x4`
    hostId = "111a0e06";
  };

  services = {
    openssh.enable = true;
    k3s = {
      enable = true;
      role = "server";
      extraFlags = toString [
        "--disable=traefik"
        "--disable=local-storage"
      ];
    };
  };

  programs.git = {
    enable = true;
    config = {
      safe = {
        directory = [ "/zfs" ];
      };
    };
  };
  programs.nix-ld.enable = true;

  systemd.tmpfiles.rules = [
    "L+ /usr/local/bin - - - - /run/current-system/sw/bin/"
  ];

  environment = {
    variables = {
      KUBECONFIG = "/etc/rancher/k3s/k3s.yaml";
      CI_K3S = "true";
      GOPATH = "/zfs/nix/.go";
      EDITOR = "vim";
      TERMINFO = "/terminfo";
    };

    shellAliases = {
      k = "kubectl";
      ke = "kubectl -n openebs";
    };

    shellInit = ''
      export PATH=$GOPATH/bin:$PATH
      cd /zfs
    '';

    systemPackages = with pkgs; [ vim docker-client k9s e2fsprogs xfsprogs ];
  };
}
