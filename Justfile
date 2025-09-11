update bindings:
    #!/usr/bin/env bash
    set -euxo pipefail
    VERSION=1.56.2

    TARGET="libguestfs.org"
    MARKER="$TARGET/.version"

    # Skip if already installed
    [[ -f "$MARKER" && "$(cat "$MARKER")" == "$VERSION" ]] && {
        echo "libguestfs Go bindings version $VERSION already present."
        exit 0
    }

    echo "Installing libguestfs Go bindings version $VERSION..."

    rm -rf "$TARGET"
    mkdir -p "$TARGET"

    # Download and extract in one step
    curl -sSL "https://download.libguestfs.org/${VERSION%.*}-stable/libguestfs-${VERSION}.tar.gz" \
        | tar --strip-components=3 -xzvf - "libguestfs-${VERSION}/golang/src/libguestfs.org" -C "$TARGET"

    # Write marker
    echo "$VERSION" > "$MARKER"

    rm $TARGET/guestfs/.gitignore $TARGET/guestfs/go.mod

    echo "Installed to $TARGET"
