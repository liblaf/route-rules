import route_rules.container.optim._utils as u


def merge_domain_with_keyword(
    domain: set[str], domain_keyword: set[str]
) -> tuple[set[str], set[str]]:
    domain_new: set[str] = set()
    keyword_new: set[str] = set(domain_keyword)
    for d in domain:
        if u.match_domain_keyword(d, keyword_new):
            continue
        domain_new.add(d)
    return domain_new, keyword_new
