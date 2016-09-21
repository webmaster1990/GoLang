import sys
import os
import inspect
import re

from docutils import nodes
import sphinx.ext.autodoc
import sphinx

sys.path.insert(0, os.path.abspath('.'))  # for Pygments Solarized style
#sys.path.insert(0, os.path.abspath('../..'))

#import api_srv

needs_sphinx = '1.2'

extensions = [
    'sphinx.ext.intersphinx',
    'sphinxcontrib.httpdomain'
]

#templates_path = ['_templates']
source_suffix = '.rst'

# The master toctree document.
master_doc = 'index'

# General information about the project.
project = 'Lucid Web App Internal API'
copyright = 'TO CHANGE'

# The short X.Y version.
version = '1.x.x'
# The full version, including alpha/beta/rc tags.
release = '1.x.x'

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
exclude_patterns = []

# default_role = None

pygments_style = 'pygments_solarized_light.LightStyle'

# A list of ignored prefixes for module index sorting.
# modindex_common_prefix = []

# -- Options for HTML output ----------------------------------------------

html_context = {'sphinx_versioninfo': sphinx.version_info}

html_theme_path = ['../theme']
html_theme = 'bootstrap'
html_theme_options = {
    'navbar_site_name': 'Lucid Web App API Docs',  # tab name for entire site. (Default: "Site")
    'navbar_sidebarrel': False,  # render the next/previous page links in navbar. (Default: True)
    #'navbar_pagenav': True,  # render the current pages TOC in the navbar. (Default: true)
    'navbar_pagenav_name': 'Page',  # tab name for the current pages TOC. (Default: "Page")
    'globaltoc_depth': 2,  # global TOC depth for "site" navbar tab (default: 1) (-1 shows all)
    #'globaltoc_includehidden': 'true',  # include hidden TOC in site navbar (default: 'true')
    #'navbar_class': 'navbar', navbar class (default: 'navbar') (opts: 'navbar navbar-inverse')
    'navbar_fixed_top': 'true',  # fixed navbar at top of page (default: 'true')
    'source_link_position': 'nav',  # location of link to source (default: nav)
    'bootswatch_theme': 'flatly',  # bootswatch theme (default: '')
}

#html_static_path = ['_static']

# Custom sidebar templates, maps document names to template names.
# html_sidebars = {}

# Output file base name for HTML help builder.
htmlhelp_basename = 'lucid_api_doc'
