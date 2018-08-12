#!/usr/bin/env python3
import gi
import sys
gi.require_version('Gtk', '3.0')
from gi.repository import Gtk

# icon_name = input("Icon name (case sensitive): ")
try:
	icon_name = sys.argv[1]
	icon_theme = Gtk.IconTheme.get_default()
	icon = icon_theme.lookup_icon(icon_name, 24, 0)
	if icon:
	    print(icon.get_filename())
	else:
	    print("not found")
except IndexError:
	print("Supply icon name as a command-line argument")