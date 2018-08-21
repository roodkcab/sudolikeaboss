#include <objc/runtime.h>
#include <Foundation/Foundation.h>
#include <AppKit/AppKit.h>
#import <mach-o/loader.h>
#import <mach-o/dyld.h>
#import <mach-o/arch.h>
#include "rd_route.h"

static int _OPVerifyAsSafariClientFaker(int a, int b)
{
    return 1;
}

@interface BYPASSSIGN : NSObject
@end

@implementation BYPASSSIGN

+ (void)load
{
    void *(*original)() = NULL;
    NSMutableDictionary *addrSlides = [NSMutableDictionary dictionaryWithDictionary:@{}];
    for (uint32_t i = 0; i < _dyld_image_count(); i++) {
            NSString *path = [NSString stringWithFormat:@"%s", _dyld_get_image_name(i)];
            NSString *name = [[path componentsSeparatedByString:@"/"] lastObject];
            [addrSlides setObject:@(_dyld_get_image_vmaddr_slide(i)) forKey:name];
    }
    unsigned long _OPVerifyAsSafariClient = ([addrSlides[@"OnePasswordCore"] unsignedLongValue] + 0x3dc3);
    rd_route((void *)_OPVerifyAsSafariClient, _OPVerifyAsSafariClientFaker, (void **)&original);
}

@end
