# Packages the binary that hack/rpm.sh builds with `make build`; the
# script passes the version in deck_version.
%global debug_package %{nil}
# Snapshots carry no %%changelog to date the build from.
%global source_date_epoch_from_changelog 0

Name:           musicfox-deck
Version:        %{deck_version}
Release:        1%{?dist}
Summary:        Netease Cloud Music in the terminal, with a 3D cassette deck
License:        GPL-3.0-only
URL:            https://github.com/Egdon/musicfox-deck
Source0:        musicfox
Source1:        LICENSE
Source2:        README.md
ExclusiveArch:  x86_64

# rpmbuild finds libFLAC, libasound and glibc in the binary itself.
# The web login window loads WebKitGTK when it opens; without it, the QR
# code login remains.
Recommends:     (webkitgtk6.0 or webkit2gtk4.1)
# Only for [player] engine = "mpv"; the default engine needs nothing.
Suggests:       mpv
# Both install the musicfox command.
Conflicts:      go-musicfox

%description
musicfox-deck is a fork of go-musicfox, a Netease Cloud Music client for
the terminal. Its main screen pairs a tracklist with a 3D cassette that
plays along, the now-playing page puts the cassette on a stage, and the
interface moves with quiet transitions. It shares the go-musicfox
configuration and data directory.

%prep
cp -p %{SOURCE1} %{SOURCE2} .

%install
install -Dpm 0755 %{SOURCE0} %{buildroot}%{_bindir}/musicfox

%files
%license LICENSE
%doc README.md
%{_bindir}/musicfox
