import ipaddress
import itertools
import re
from collections.abc import Iterable


def _split_domain(domain: str) -> list[str]:
    return re.split(r"(\.)", domain)


def _domain_suffixes(domain: str) -> list[str]:
    labels: list[str] = _split_domain(domain[::-1])
    suffixes: list[str] = [suffix[::-1] for suffix in itertools.accumulate(labels)]
    return suffixes


def _match_domain_suffix(domain: str, suffix: set[str]) -> bool:
    return any((s in suffix) for s in _domain_suffixes(domain))


def _match_domain_keyword(domain: str, keyword: set[str]) -> bool:
    return any((k in domain) for k in keyword)


def merge_domain_with_suffix(
    domain: Iterable[str], domain_suffix: Iterable[str]
) -> tuple[set[str], set[str]]:
    domain_result: set[str] = set()
    suffix: set[str] = set(domain_suffix)
    for d in domain:
        if "." + d in suffix:
            suffix.remove("." + d)
            suffix.add(d)
            continue
        if _match_domain_suffix(d, suffix):
            continue
        domain_result.add(d)
    return domain_result, suffix


def merge_between_domain_suffix(domain_suffix: Iterable[str]) -> set[str]:
    domain_suffix = sorted(domain_suffix, key=len)
    suffix: set[str] = set()
    for s in domain_suffix:
        if _match_domain_suffix(s, suffix):
            continue
        suffix.add(s)
    return suffix


def merge_domain_with_keyword(
    domain: Iterable[str], domain_keyword: Iterable[str]
) -> tuple[set[str], set[str]]:
    domain_result: set[str] = set()
    keyword: set[str] = set(domain_keyword)
    for d in domain:
        if _match_domain_keyword(d, keyword):
            continue
        domain_result.add(d)
    return domain_result, keyword


def merge_domain_suffix_with_keyword(
    domain_suffix: Iterable[str], domain_keyword: Iterable[str]
) -> tuple[set[str], set[str]]:
    suffix: set[str] = set()
    keyword: set[str] = set(domain_keyword)
    for d in domain_suffix:
        if _match_domain_keyword(d, keyword):
            continue
        suffix.add(d)
    return suffix, keyword


def merge_ip_cidr(ip_cidr: Iterable[str]) -> list[str]:
    ipv4_networks: list[ipaddress.IPv4Network] = []
    ipv6_networks: list[ipaddress.IPv6Network] = []
    for cidr in ip_cidr:
        network: ipaddress.IPv4Network | ipaddress.IPv6Network = ipaddress.ip_network(
            cidr
        )
        if network.version == 4:
            ipv4_networks.append(network)
        elif network.version == 6:
            ipv6_networks.append(network)
    ipv4_networks = list(ipaddress.collapse_addresses(ipv4_networks))
    ipv6_networks = list(ipaddress.collapse_addresses(ipv6_networks))
    result: list[str] = [
        str(network)
        for network in itertools.chain(
            ipaddress.collapse_addresses(ipv4_networks),
            ipaddress.collapse_addresses(ipv6_networks),
        )
    ]
    return result
