# -*- python -*-

# stdlib imports ---
import os
import os.path as osp
import subprocess

# waf imports ---
import waflib.Utils
import waflib.Logs as msg
from waflib.Configure import conf

#
_heptooldir = osp.dirname(osp.abspath(__file__))

def options(ctx):

    ctx.load('hwaf-base', tooldir=_heptooldir)

    ctx.load('boost')
    ctx.add_option(
        '--with-boost',
        default=None,
        help="Look for boost at the given path")
    return

def configure(ctx):
    ctx.load('hwaf-base', tooldir=_heptooldir)
    return

@conf
def find_boost(ctx, **kwargs):
    
    ctx.load('hwaf-base', tooldir=_heptooldir)

    if not ctx.env.HWAF_FOUND_C_COMPILER:
        ctx.fatal('load a C compiler first')
        pass

    if not ctx.env.HWAF_FOUND_CXX_COMPILER:
        ctx.fatal('load a C++ compiler first')
        pass

    if not ctx.env.HWAF_FOUND_PYTHON:
        ctx.find_python()
        pass

    xx,yy = ctx.env.PYTHON_VERSION.split(".")[:2]
    ctx.options.boost_python = "%s%s" % (xx,yy)
    
    ctx.load('boost')
    boost_libs = '''\
    chrono date_time filesystem graph iostreams
    math_c99 math_c99f math_tr1 math_tr1f
    prg_exec_monitor program_options 
    random regex serialization
    signals system thread 
    unit_test_framework wave wserialization
    '''

    kwargs['mt'] = kwargs.get('mt', False)
    kwargs['static'] = kwargs.get('static', False)
    kwargs['use'] = waflib.Utils.to_list(kwargs.get('use', [])) + ['python']

    ctx.check_with(
        ctx.check_boost,
        "boost",
        lib=boost_libs,
        uselib_store='boost',
        **kwargs)

    for libname in boost_libs.split():
        libname = libname.strip()
        for n in ('INCLUDES',
                  'LIBPATH',
                  'LINKFLAGS'):
            ctx.env['%s_boost-%s'%(n,libname)] = ctx.env['%s_boost'%n][:]
        lib = []
        for i in ctx.env['LIB_boost']:
            if i.startswith("boost_%s-"%libname):
                lib.append(i)
                break
            if i == "boost_%s"%libname:
                lib.append(i)
                break
        else:
            msg.warn("could not find a linkopt for [boost_%s] among: %s" %
                     (libname,ctx.env['LIB_boost']))
        ctx.env['LIB_boost-%s'%libname] = lib[:]

    # register the boost module
    import sys
    modname = 'waflib.extras.boost'
    fname = sys.modules[modname].__file__
    if fname.endswith('.pyc'): fname = fname[:-1]
    ctx.hwaf_export_module(ctx.root.find_node(fname).abspath())

    ctx.env.HWAF_FOUND_BOOST = 1
    return


