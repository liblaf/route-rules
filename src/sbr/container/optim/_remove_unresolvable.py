import validators


def remove_unresolvable(domain: set[str]) -> set[str]:
    return {d for d in domain if validators.domain(d)}
