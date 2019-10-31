%define revision 1

%global provider        github
%global provider_tld    com
%global project         lazyfrosch
%global repo            filespooler
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}

%global commit          381ca4658e0edca1fac2345543c097ac9e547202
%global shortcommit     %(c=%{commit}; echo ${c:0:7})

%global golang_min_version 1.11

%global daemon_user     icinga
%global daemon_group    icinga

%global service_receiver filespooler-receiver
%global service_sender   filespooler-sender

%global golang_pkg      golang
#%%if 0%{?el7:1}
#%%global golang_scl      go-toolset-X
#%%endif # el7

%if 0%{?golang_scl:1}
%global golang_scl_prefix %{golang_scl}-
%global golang_scl_enable scl enable %{golang_scl} --
%endif

%if ! 0%{?gobuild:1}
%define gobuild(o:) %{?golang_scl_enable} go build -ldflags "${LDFLAGS:-} -B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \\n')" -a -v -x %{?**};
%endif

%if ! 0%{?gopath:1}
%define gopath %(%{?golang_scl_enable} go env GOPATH)
%endif

Name:       filespooler
Version:    0.0.0.%{shortcommit}
Release:    %{revision}%{?dist}
Summary:    NETWAYS File Spooler
Group:      System Environment/Daemons
License:    GPLv2+
URL:        https://%{provider_prefix}
Source0:    https://%{import_path}/archive/%{commit}.tar.gz#/%{repo}-%{version}.tar.gz

BuildRoot:  %{_tmppath}/%{name}-%{version}-%{release}

BuildRequires:  %{?golang_scl_prefix}%{golang_pkg} >= %{golang_min_version}

%{?systemd_requires}
BuildRequires:  systemd

Requires(pre):  shadow-utils

%description
NETWAYS File Spooler to transport spooled files between servers.

%prep
%setup -q -n %{name}-%{commit}

%build
%gobuild ./cmd/filespooler

%check
%if ! 0%{?gotest:1}
%global gotest %{?golang_scl_enable} go test
%endif

%gotest ./...

%install
install -d -m 0755 %{buildroot}%{_sbindir}
install -d -m 0755 %{buildroot}%{_unitdir}

install -m 0755 filespooler %{buildroot}%{_sbindir}/%{name}
#install -m 0644 systemd/filespooler-receiver.service %%{buildroot}%%{_unitdir}/%%{service_receiver}.service
#install -m 0644 systemd/filespooler-sender.service %%{buildroot}%%{_unitdir}/%%{service_sender}.service

%pre
getent group %{daemon_group} >/dev/null || groupadd -r %{daemon_group}
getent passwd %{daemon_user} >/dev/null || useradd -r -g %{daemon_group} -d / -s /sbin/nologin %{daemon_user}

#%post
#%systemd_post %{service_receiver}
#%systemd_post %{service_sender}

#%preun
#%systemd_preun %{service_receiver}
#%systemd_preun %{service_sender}

#%postun
#%systemd_postun %{service_receiver}
#%systemd_postun %{service_sender}

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root)
%doc COPYING README.md
%{_sbindir}/%{name}
#%%{_unitdir}/%{service_receiver}.service
#%%{_unitdir}/%{service_sender}.service

%defattr(0640,%{daemon_user},%{daemon_group},0750)

%changelog
* Mon Oct 14 2019 Markus Frosch <markus.frosch@netways.de> 0.0.0-0
- Initial package
