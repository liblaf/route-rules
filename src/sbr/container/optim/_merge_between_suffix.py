import sbr.container.optim._utils as u


def merge_between_suffix(domain_suffix: set[str]) -> set[str]:
    suffix_list: list[str] = sorted(domain_suffix, key=len)
    suffix: set[str] = set()
    for s in suffix_list:
        if u.match_domain_suffix(s, suffix):
            continue
        suffix.add(s)
    return suffix
