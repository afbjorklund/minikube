CONMON_MASTER_VERSION = dde3ccf93f01ce5a3e0f7a2c97053697cc3ed152
CONMON_MASTER_SITE = https://github.com/containers/conmon/archive
CONMON_MASTER_SOURCE = $(CONMON_MASTER_VERSION).tar.gz
CONMON_MASTER_LICENSE = Apache-2.0
CONMON_MASTER_LICENSE_FILES = LICENSE

CONMON_MASTER_DEPENDENCIES = host-pkgconf

define CONMON_MASTER_PATCH_PKGCONFIG
	sed -e 's/pkg-config/$$(PKG_CONFIG)/g' -i $(@D)/Makefile
endef

CONMON_MASTER_POST_PATCH_HOOKS += CONMON_MASTER_PATCH_PKGCONFIG

define CONMON_MASTER_BUILD_CMDS
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) GIT_COMMIT=$(CONMON_MASTER_VERSION) PREFIX=/usr
endef

define CONMON_MASTER_INSTALL_TARGET_CMDS
	# crio conmon is installed by the crio package, so don't install it here
	$(INSTALL) -Dm755 $(@D)/bin/conmon $(TARGET_DIR)/usr/libexec/podman/conmon
endef

$(eval $(generic-package))
