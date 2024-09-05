import ipaddress
import itertools
from ipaddress import IPv4Network, IPv6Network


def merge_ip_cidr(ip_cidr: set[str]) -> set[str]:
    ipv4_networks: list[IPv4Network] = []
    ipv6_networks: list[IPv6Network] = []
    for cidr in ip_cidr:
        network: IPv4Network | IPv6Network = ipaddress.ip_network(cidr)
        if network.version == 4:
            ipv4_networks.append(network)
        elif network.version == 6:
            ipv6_networks.append(network)
    result: set[str] = {
        str(network)
        for network in itertools.chain(
            ipaddress.collapse_addresses(ipv4_networks),
            ipaddress.collapse_addresses(ipv6_networks),
        )
    }
    return result
